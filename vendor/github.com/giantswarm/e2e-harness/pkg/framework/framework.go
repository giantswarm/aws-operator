package framework

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	giantclientset "github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/e2e-harness/pkg/harness"
)

// PatchSpec is a generic patch type to update objects with JSONPatchType operations.
type PatchSpec struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

const (
	// minimumNodesReady represents the minimun number of ready nodes in a guest
	// cluster to be considered healthy.
	minimumNodesReady = 3

	certOperatorValuesFile = "/tmp/cert-operator-values.yaml"
	// certOperatorChartValues values required by cert-operator-chart, the environment
	// variables will be expanded before writing the contents to a file.
	certOperatorChartValues = `commonDomain: ${COMMON_DOMAIN_GUEST}
clusterName: ${CLUSTER_NAME}
Installation:
  V1:
    Auth:
      Vault:
        Address: http://vault.default.svc.cluster.local:8200
        CA:
          TTL: 1440h
    Guest:
      Kubernetes:
        API:
          EndpointBase: ${COMMON_DOMAIN_GUEST}
    Secret:
      CertOperator:
        SecretYaml: |
          service:
            vault:
              config:
                token: ${VAULT_TOKEN}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
`
)

type Framework struct {
	cs      kubernetes.Interface
	gsCs    *giantclientset.Clientset
	guestCS kubernetes.Interface
}

func New() (*Framework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	gsCs, err := giantclientset.NewForConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	f := &Framework{
		cs:   cs,
		gsCs: gsCs,
	}
	return f, nil
}

func (f *Framework) runningPod(namespace, labelSelector string) func() error {
	return func() error {
		pods, err := f.cs.CoreV1().
			Pods(namespace).
			List(metav1.ListOptions{
				LabelSelector: labelSelector,
			})
		if err != nil {
			return microerror.Mask(err)
		}
		if len(pods.Items) > 1 {
			return microerror.Mask(tooManyResultsError)
		}
		pod := pods.Items[0]
		phase := pod.Status.Phase
		if phase != v1.PodRunning {
			return microerror.Maskf(unexpectedStatusPhaseError, "current status: %s", string(phase))
		}
		return nil
	}
}

func (f *Framework) secret(namespace, secretName string) func() error {
	return func() error {
		_, err := f.cs.CoreV1().
			Secrets(namespace).
			Get(secretName, metav1.GetOptions{})
		return microerror.Mask(err)
	}
}

func (f *Framework) crd(crdName string) func() error {
	return func() error {
		// FIXME: use proper clientset call when apiextensions are in place,
		// `cs.ExtensionsV1beta1().ThirdPartyResources().Get(tprName, metav1.GetOptions{})` finding
		// the tpr is not enough for being able to create a tpo.
		return runCmd("kubectl get " + crdName)
	}
}

func (f *Framework) WaitForPodLog(namespace, needle, podName string) error {
	needle = os.ExpandEnv(needle)

	timeout := time.After(defaultTimeout * time.Second)

	req := f.cs.CoreV1().
		RESTClient().
		Get().
		Namespace(namespace).
		Name(podName).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true))

	readCloser, err := req.Stream()
	if err != nil {
		return microerror.Mask(err)
	}
	defer readCloser.Close()

	scanner := bufio.NewScanner(readCloser)
	var lastLine string
	for scanner.Scan() {
		select {
		case <-timeout:
			return microerror.Mask(waitTimeoutError)
		default:
		}
		lastLine = scanner.Text()
		log.Print(lastLine)
		if strings.Contains(lastLine, needle) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return microerror.Mask(err)
	}

	return microerror.Mask(notFoundError)
}

func (f *Framework) PodName(namespace, labelSelector string) (string, error) {
	pods, err := f.cs.CoreV1().
		Pods(namespace).
		List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(pods.Items) > 1 {
		return "", microerror.Mask(tooManyResultsError)
	}
	if len(pods.Items) == 0 {
		return "", microerror.Mask(notFoundError)
	}
	pod := pods.Items[0]
	return pod.Name, nil
}

func (f *Framework) SetUp() error {
	if err := f.createGSNamespace(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.installVault(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (f *Framework) TearDown() {
	runCmd("helm delete vault --purge")
	f.cs.CoreV1().
		Namespaces().
		Delete("giantswarm", &metav1.DeleteOptions{})
}

func (f *Framework) createGSNamespace() error {
	// check if the namespace already exists
	_, err := f.cs.CoreV1().
		Namespaces().
		Get("giantswarm", metav1.GetOptions{})
	if err == nil {
		return nil
	}

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "giantswarm",
		},
	}
	_, err = f.cs.CoreV1().
		Namespaces().
		Create(namespace)
	if err != nil {
		return microerror.Mask(err)
	}

	activeNamespace := func() error {
		ns, err := f.cs.CoreV1().
			Namespaces().
			Get("giantswarm", metav1.GetOptions{})

		if err != nil {
			return microerror.Mask(err)
		}

		phase := ns.Status.Phase
		if phase != v1.NamespaceActive {
			return microerror.Maskf(unexpectedStatusPhaseError, "current status: %s", string(phase))
		}

		return nil
	}

	return waitFor(activeNamespace)
}

func (f *Framework) installVault() error {
	if err := runCmd("helm registry install quay.io/giantswarm/vaultlab-chart:stable -- --set vaultToken=${VAULT_TOKEN} -n vault"); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.runningPod("default", "app=vault"))
}

func (f *Framework) InstallCertOperator() error {
	certOperatorChartValuesEnv := os.ExpandEnv(certOperatorChartValues)
	if err := ioutil.WriteFile(certOperatorValuesFile, []byte(certOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/cert-operator-chart:stable -- -n cert-operator --values " + certOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.crd("certconfig"))
}

func (f *Framework) InstallCertResource() error {
	err := runCmd("helm registry install quay.io/giantswarm/cert-resource-lab-chart:stable -- -n cert-resource-lab --set commonDomain=${COMMON_DOMAIN_GUEST} --set clusterName=${CLUSTER_NAME}")
	if err != nil {
		return microerror.Mask(err)
	}

	secretName := fmt.Sprintf("%s-api", os.Getenv("CLUSTER_NAME"))
	log.Printf("waiting for secret %v\n", secretName)
	return waitFor(f.secret("default", secretName))
}

func (f *Framework) InstallAwsOperator(values string) error {
	awsOperatorChartValuesEnv := os.ExpandEnv(values)

	tmpfile, err := ioutil.TempFile("", "aws-operator-values")
	if err != nil {
		return microerror.Mask(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(awsOperatorChartValuesEnv)); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/aws-operator-chart@1.0.0-${CIRCLE_SHA1} -- -n aws-operator --values " + tmpfile.Name()); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.crd("awsconfig"))
}

func (f *Framework) InstallNodeOperator(values string) error {
	var err error

	nodeOperatorChartValuesEnv := os.ExpandEnv(values)

	tmpfile, err := ioutil.TempFile("", "node-operator-values")
	if err != nil {
		return microerror.Mask(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(nodeOperatorChartValuesEnv)); err != nil {
		return microerror.Mask(err)
	}

	err = runCmd("helm registry install quay.io/giantswarm/node-operator-chart:stable -- -n node-operator --values " + tmpfile.Name())
	if err != nil {
		return microerror.Mask(err)
	}

	err = waitFor(f.crd("nodeconfig"))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (f *Framework) DeleteGuestCluster() error {
	if err := runCmd("kubectl delete awsconfig ${CLUSTER_NAME}"); err != nil {
		return microerror.Mask(err)
	}

	operatorPodName, err := f.PodName("giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	logEntry := "deleted the guest cluster main stack"

	return f.WaitForPodLog("giantswarm", logEntry, operatorPodName)
}

func (f *Framework) initGuestClientset() error {
	if f.guestCS != nil {
		return nil
	}
	// get api secret
	secretName := os.ExpandEnv("${CLUSTER_NAME}-api")

	secret, err := f.cs.CoreV1().
		Secrets("default").
		Get(secretName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	// create clientset
	config := &rest.Config{}
	config.TLSClientConfig = rest.TLSClientConfig{
		CAData:   secret.Data["ca"],
		CertData: secret.Data["crt"],
		KeyData:  secret.Data["key"],
	}
	config.Host = os.ExpandEnv("https://api.${CLUSTER_NAME}.${COMMON_DOMAIN_GUEST}")

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return microerror.Mask(err)
	}

	f.guestCS = cs

	return nil
}

func (f *Framework) WaitForGuestReady() error {
	if err := f.initGuestClientset(); err != nil {
		return microerror.Maskf(err, "unexpected error initializing guest clientset")
	}

	if err := f.waitForAPIUp(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.WaitForNodesUp(minimumNodesReady); err != nil {
		return microerror.Mask(err)
	}
	log.Println("Guest cluster ready")
	return nil
}

func (f *Framework) WaitForNodesUp(numberOfNodes int) error {
	nodesUp := func() error {
		res, err := f.guestCS.
			CoreV1().
			Nodes().
			List(metav1.ListOptions{})

		if err != nil {
			log.Printf("waiting for nodes ready: %v\n", err)
			return microerror.Mask(err)
		}
		if len(res.Items) != numberOfNodes {
			log.Printf("worker nodes not found")
			return microerror.Mask(notFoundError)
		}

		for _, n := range res.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == v1.NodeReady && c.Status != v1.ConditionTrue {
					log.Printf("not all worker nodes ready")
					return microerror.Mask(notFoundError)
				}
			}
		}
		return nil
	}

	return waitFor(nodesUp)
}

func (f *Framework) waitForAPIUp() error {
	apiUp := func() error {
		_, err := f.guestCS.
			CoreV1().
			Services("default").
			Get("kubernetes", metav1.GetOptions{})

		if err != nil {
			log.Printf("waiting for k8s API up: %v\n", err)
			return microerror.Mask(err)
		}
		log.Println("k8s API up")
		return nil
	}

	return waitFor(apiUp)
}

func (f *Framework) WaitForAPIDown() error {
	apiDown := func() error {
		_, err := f.guestCS.
			CoreV1().
			Services("default").
			Get("kubernetes", metav1.GetOptions{})

		if err == nil {
			return microerror.Mask(fmt.Errorf("API up"))
		}
		log.Printf("k8s API down: %v\n", err)
		return nil
	}

	log.Printf("waiting for k8s API down\n")
	return waitConstantFor(apiDown)
}

func (f *Framework) AWSCluster(name string) (*v1alpha1.AWSConfig, error) {
	cluster, err := f.gsCs.ProviderV1alpha1().
		AWSConfigs("default").
		Get(name, metav1.GetOptions{})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cluster, nil
}

func (f *Framework) ApplyAWSConfigPatch(patch []PatchSpec, clusterName string) error {
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = f.gsCs.
		ProviderV1alpha1().
		AWSConfigs("default").
		Patch(clusterName, types.JSONPatchType, patchBytes)

	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func GetVersionBundleVersion(bundle []versionbundle.Bundle, vType string) (string, error) {
	validVTypes := []string{"", "current", "wip"}
	var isValid bool
	for _, v := range validVTypes {
		if v == vType {
			isValid = true
			break
		}
	}
	if !isValid {
		return "", fmt.Errorf("%q is not a valid version bundle version type", vType)
	}

	var output string
	log.Printf("Tested version %q", vType)

	// sort bundle by time to get the newest vbv.
	s := versionbundle.SortBundlesByTime(bundle)
	sort.Sort(sort.Reverse(s))
	for _, v := range s {
		if (vType == "current" || vType == "") && !v.Deprecated && !v.WIP {
			output = v.Version
			break
		}
		if vType == "wip" && v.WIP {
			output = v.Version
			break
		}
	}
	log.Printf("Version Bundle Version %q", output)
	return output, nil
}

// HelmCmd executes a helm command.
func HelmCmd(cmd string) error {
	return runCmd("helm " + cmd)
}
