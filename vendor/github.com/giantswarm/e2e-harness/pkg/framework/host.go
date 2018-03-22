package framework

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	giantclientset "github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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

type Host struct {
	backoff   *backoff.ExponentialBackOff
	g8sClient *giantclientset.Clientset
	k8sClient kubernetes.Interface
}

type Config struct {
	Backoff *backoff.ExponentialBackOff
}

func NewHost(c *Config) (*Host, error) {
	if c.Backoff == nil {
		c.Backoff = newCustomExponentialBackoff()
	}

	config, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := giantclientset.NewForConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	h := &Host{
		backoff:   c.Backoff,
		g8sClient: g8sClient,
		k8sClient: k8sClient,
	}

	return h, nil
}

func (h *Host) ApplyAWSConfigPatch(patch []PatchSpec, clusterName string) error {
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = h.g8sClient.
		ProviderV1alpha1().
		AWSConfigs("default").
		Patch(clusterName, types.JSONPatchType, patchBytes)

	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *Host) AWSCluster(name string) (*v1alpha1.AWSConfig, error) {
	cluster, err := h.g8sClient.ProviderV1alpha1().
		AWSConfigs("default").
		Get(name, metav1.GetOptions{})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cluster, nil
}

func (h *Host) DeleteGuestCluster() error {
	if err := runCmd("kubectl delete awsconfig ${CLUSTER_NAME}"); err != nil {
		return microerror.Mask(err)
	}

	operatorPodName, err := h.PodName("giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	logEntry := "deleted the guest cluster main stack"

	return h.WaitForPodLog("giantswarm", logEntry, operatorPodName)
}

func (h *Host) InstallAwsOperator(values string) error {
	awsOperatorChartValuesEnv := os.ExpandEnv(values)

	tmpfile, err := ioutil.TempFile("", "aws-operator-values")
	if err != nil {
		return microerror.Mask(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(awsOperatorChartValuesEnv)); err != nil {
		return microerror.Mask(err)
	}

	operation := func() error {
		return HelmCmd("registry install quay.io/giantswarm/aws-operator-chart@1.0.0-${CIRCLE_SHA1} -- -n aws-operator --values " + tmpfile.Name())
	}
	notify := newNotify("aws-operator-chart install")
	err = backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	return waitFor(h.crd("awsconfig"))
}

func (h *Host) InstallCertOperator() error {
	certOperatorChartValuesEnv := os.ExpandEnv(certOperatorChartValues)
	if err := ioutil.WriteFile(certOperatorValuesFile, []byte(certOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	operation := func() error {
		return HelmCmd("registry install quay.io/giantswarm/cert-operator-chart:stable -- -n cert-operator --values " + certOperatorValuesFile)
	}
	notify := newNotify("cert-operator-chart install")
	err := backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	return waitFor(h.crd("certconfig"))
}

func (h *Host) InstallCertResource() error {
	operation := func() error {
		return HelmCmd("registry install quay.io/giantswarm/cert-resource-lab-chart:stable -- -n cert-resource-lab --set commonDomain=${COMMON_DOMAIN_GUEST} --set clusterName=${CLUSTER_NAME}")
	}
	notify := newNotify("cert-resource-lab-chart install")
	err := backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	secretName := fmt.Sprintf("%s-api", os.Getenv("CLUSTER_NAME"))
	log.Printf("waiting for secret %v\n", secretName)
	return waitFor(h.secret("default", secretName))
}

func (h *Host) InstallNodeOperator(values string) error {
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

	operation := func() error {
		return HelmCmd("registry install quay.io/giantswarm/node-operator-chart:stable -- -n node-operator --values " + tmpfile.Name())
	}
	notify := newNotify("node-operator-chart install")
	err = backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	err = waitFor(h.crd("nodeconfig"))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *Host) PodName(namespace, labelSelector string) (string, error) {
	pods, err := h.k8sClient.CoreV1().
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

func (h *Host) Setup() error {
	if err := h.createGSNamespace(); err != nil {
		return microerror.Mask(err)
	}

	if err := h.installVault(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *Host) Teardown() {
	HelmCmd("delete vault --purge")
	h.k8sClient.CoreV1().
		Namespaces().
		Delete("giantswarm", &metav1.DeleteOptions{})
}

func (h *Host) WaitForPodLog(namespace, needle, podName string) error {
	needle = os.ExpandEnv(needle)

	timeout := time.After(defaultTimeout * time.Second)

	req := h.k8sClient.CoreV1().
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

func (h *Host) crd(crdName string) func() error {
	return func() error {
		// FIXME: use proper clientset call when apiextensions are in place,
		// `k8sClient.ExtensionsV1beta1().ThirdPartyResources().Get(tprName, metav1.GetOptions{})` finding
		// the tpr is not enough for being able to create a tpo.
		return runCmd("kubectl get " + crdName)
	}
}

func (h *Host) createGSNamespace() error {
	// check if the namespace already exists
	_, err := h.k8sClient.CoreV1().
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
	_, err = h.k8sClient.CoreV1().
		Namespaces().
		Create(namespace)
	if err != nil {
		return microerror.Mask(err)
	}

	activeNamespace := func() error {
		ns, err := h.k8sClient.CoreV1().
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

func (h *Host) installVault() error {
	operation := func() error {
		return HelmCmd("registry install quay.io/giantswarm/vaultlab-chart:stable -- --set vaultToken=${VAULT_TOKEN} -n vault")
	}
	notify := newNotify("vaultlab-chart install")
	err := backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	return waitFor(h.runningPod("default", "app=vault"))
}

func (h *Host) secret(namespace, secretName string) func() error {
	return func() error {
		_, err := h.k8sClient.CoreV1().
			Secrets(namespace).
			Get(secretName, metav1.GetOptions{})
		return microerror.Mask(err)
	}
}

func (h *Host) runningPod(namespace, labelSelector string) func() error {
	return func() error {
		pods, err := h.k8sClient.CoreV1().
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
