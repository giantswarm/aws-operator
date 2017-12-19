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
	cs   kubernetes.Interface
	gsCs *giantclientset.Clientset
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

func (f *framework) waitForPodLog(namespace, needle, podName string) error {
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

func (f *framework) podName(namespace, labelSelector string) (string, error) {
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

func (f *framework) setUp() error {
	if err := f.createGSNamespace(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.installVault(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (f *framework) tearDown() {
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

func (f *framework) installCertOperator() error {
	certOperatorChartValuesEnv := os.ExpandEnv(certOperatorChartValues)
	if err := ioutil.WriteFile(certOperatorValuesFile, []byte(certOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/cert-operator-chart:stable -- -n cert-operator --values " + certOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.crd("certconfig"))
}

func (f *framework) installCertResource() error {
	err := runCmd("helm registry install quay.io/giantswarm/cert-resource-lab-chart:stable -- -n cert-resource-lab --set commonDomain=${COMMON_DOMAIN} --set clusterName=${CLUSTER_NAME}")
	if err != nil {
		return microerror.Mask(err)
	}

	secretName := fmt.Sprintf("%s-api", os.Getenv("CLUSTER_NAME"))
	log.Printf("waiting for secret %v\n", secretName)
	return waitFor(f.secret("default", secretName))
}

func (f *framework) installAwsOperator() error {
	awsOperatorChartValuesEnv := os.ExpandEnv(awsOperatorChartValues)
	if err := ioutil.WriteFile(awsOperatorValuesFile, []byte(awsOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/aws-operator-chart@1.0.0-${CIRCLE_SHA1} -- -n aws-operator --values " + awsOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(f.crd("awsconfig"))
}

func (f *framework) deleteGuestCluster() error {
	if err := runCmd("kubectl delete awsconfig ${CLUSTER_NAME}"); err != nil {
		return microerror.Mask(err)
	}

	operatorPodName, err := f.podName("giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	logEntry := "cluster '${CLUSTER_NAME}' deleted"
	if os.Getenv("VERSION_BUNDLE_VERSION") == "0.2.0" {
		logEntry = "deleting AWS cloudformation stack: deleted"
	}
	return f.waitForPodLog("giantswarm", logEntry, operatorPodName)
}
