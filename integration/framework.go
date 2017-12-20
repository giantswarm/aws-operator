// +build k8srequired

package integration

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	giantclientset "github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	certOperatorValuesFile = "/tmp/cert-operator-values.yaml"
	// certOperatorChartValues values required by cert-operator-chart, the environment
	// variables will be expanded before writing the contents to a file.
	certOperatorChartValues = `commonDomain: ${COMMON_DOMAIN}
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
          EndpointBase: ${COMMON_DOMAIN}
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

type framework struct {
	cs      kubernetes.Interface
	gsCs    *giantclientset.Clientset
	guestCS kubernetes.Interface
}

func newFramework() (*framework, error) {
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

	f := &framework{
		cs:   cs,
		gsCs: gsCs,
	}
	return f, nil
}

func (f *framework) runningPod(namespace, labelSelector string) func() error {
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

func (f *framework) activeNamespace(name string) func() error {
	return func() error {
		ns, err := f.cs.CoreV1().
			Namespaces().
			Get(name, metav1.GetOptions{})

		if err != nil {
			return microerror.Mask(err)
		}

		phase := ns.Status.Phase
		if phase != v1.NamespaceActive {
			return microerror.Maskf(unexpectedStatusPhaseError, "current status: %s", string(phase))
		}

		return nil
	}
}

func (f *framework) secret(namespace, secretName string) func() error {
	return func() error {
		_, err := f.cs.CoreV1().
			Secrets(namespace).
			Get(secretName, metav1.GetOptions{})
		return microerror.Mask(err)
	}
}

func (f *framework) crd(crdName string) func() error {
	return func() error {
		// FIXME: use proper clientset call when apiextensions are in place,
		// `cs.ExtensionsV1beta1().ThirdPartyResources().Get(tprName, metav1.GetOptions{})` finding
		// the tpr is not enough for being able to create a tpo.
		return runCmd("kubectl get " + crdName)
	}
}

func (f *framework) WaitForPodLog(namespace, needle, podName string) error {
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

func (f *framework) PodName(namespace, labelSelector string) (string, error) {
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

func (f *framework) SetUp() error {
	if err := f.createGSNamespace(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.installVault(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (f *framework) TearDown() {
	runCmd("helm delete vault --purge")
	f.cs.CoreV1().
		Namespaces().
		Delete("giantswarm", &metav1.DeleteOptions{})
}

func (f *framework) createGSNamespace() error {
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

	return waitFor(f.activeNamespace("giantswarm"))
}

func (f *framework) installVault() error {
	if err := runCmd("helm registry install quay.io/giantswarm/vaultlab-chart:stable -- --set vaultToken=${VAULT_TOKEN} -n vault"); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.runningPod("default", "app=vault"))
}

func (f *framework) InstallCertOperator() error {
	certOperatorChartValuesEnv := os.ExpandEnv(certOperatorChartValues)
	if err := ioutil.WriteFile(certOperatorValuesFile, []byte(certOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/cert-operator-chart:stable -- -n cert-operator --values " + certOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.crd("certconfig"))
}

func (f *framework) InstallCertResource() error {
	err := runCmd("helm registry install quay.io/giantswarm/cert-resource-lab-chart:stable -- -n cert-resource-lab --set commonDomain=${COMMON_DOMAIN} --set clusterName=${CLUSTER_NAME}")
	if err != nil {
		return microerror.Mask(err)
	}

	secretName := fmt.Sprintf("%s-api", os.Getenv("CLUSTER_NAME"))
	log.Printf("waiting for secret %v\n", secretName)
	return waitFor(f.secret("default", secretName))
}

func (f *framework) InstallAwsOperator() error {
	awsOperatorChartValuesEnv := os.ExpandEnv(awsOperatorChartValues)
	if err := ioutil.WriteFile(awsOperatorValuesFile, []byte(awsOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/aws-operator-chart@1.0.0-${CIRCLE_SHA1} -- -n aws-operator --values " + awsOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.crd("awsconfig"))
}

func (f *framework) DeleteGuestCluster() error {
	if err := runCmd("kubectl delete awsconfig ${CLUSTER_NAME}"); err != nil {
		return microerror.Mask(err)
	}

	operatorPodName, err := f.PodName("giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	// TODO: during the cloudformation migration the legacy resource is always deleted last,
	// when the migration is done we will need to check here the cloudformation stack deletion
	// message
	logEntry := "cluster '${CLUSTER_NAME}' deleted"
	return f.WaitForPodLog("giantswarm", logEntry, operatorPodName)
}

func (f *framework) initGuestClientset() error {
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
	config.Host = os.ExpandEnv("https://api.${CLUSTER_NAME}.${COMMON_DOMAIN}")

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return microerror.Mask(err)
	}

	f.guestCS = cs

	return nil
}

func (f *framework) WaitForAPIUp() error {
	if err := f.initGuestClientset(); err != nil {
		return microerror.Maskf(err, "unexpected error initializing guest clientset")
	}

	return waitFor(f.apiUp())
}

func (f *framework) apiUp() func() error {
	return func() error {
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
}
