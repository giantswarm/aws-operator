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
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
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

type HostConfig struct {
	Backoff *backoff.ExponentialBackOff
}

type Host struct {
	backoff    *backoff.ExponentialBackOff
	g8sClient  *versioned.Clientset
	k8sClient  kubernetes.Interface
	restConfig *rest.Config
}

func NewHost(c HostConfig) (*Host, error) {
	if c.Backoff == nil {
		c.Backoff = newCustomExponentialBackoff()
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	h := &Host{
		backoff:    c.Backoff,
		g8sClient:  g8sClient,
		k8sClient:  k8sClient,
		restConfig: restConfig,
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

func (h *Host) CreateNamespace(ns string) error {
	// check if the namespace already exists
	_, err := h.k8sClient.CoreV1().
		Namespaces().
		Get(ns, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
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
			Get(ns, metav1.GetOptions{})

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

func (h *Host) DeleteGuestCluster(name, cr, logEntry string) error {
	if err := runCmd(fmt.Sprintf("kubectl delete %s ${CLUSTER_NAME}", cr)); err != nil {
		return microerror.Mask(err)
	}

	operatorPodName, err := h.PodName("giantswarm", fmt.Sprintf("app=%s", name))
	if err != nil {
		return microerror.Mask(err)
	}

	return h.WaitForPodLog("giantswarm", logEntry, operatorPodName)
}

// G8sClient returns the host cluster framework's Giant Swarm client.
func (h *Host) G8sClient() versioned.Interface {
	return h.g8sClient
}

func (h *Host) GetPodName(namespace, labelSelector string) (string, error) {
	o := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pods, err := h.k8sClient.CoreV1().Pods(namespace).List(o)
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

func (h *Host) InstallStableOperator(name, cr, values string) error {
	err := h.InstallOperator(name, cr, values, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (h *Host) InstallBranchOperator(name, cr, values string) error {
	err := h.InstallOperator(name, cr, values, "@1.0.0-${CIRCLE_SHA1}")
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (h *Host) InstallOperator(name, cr, values, version string) error {
	return h.InstallResource(name, values, version, h.crd(cr))
}

func (h *Host) InstallResource(name, values, version string, conditions ...func() error) error {
	chartValuesEnv := os.ExpandEnv(values)

	tmpfile, err := ioutil.TempFile("", name+"-values")
	if err != nil {
		return microerror.Mask(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(chartValuesEnv)); err != nil {
		return microerror.Mask(err)
	}

	installCmd := fmt.Sprintf("registry install quay.io/giantswarm/%[1]s-chart%[2]s -- -n %[1]s --values %[3]s", name, version, tmpfile.Name())
	deleteCmd := fmt.Sprintf("delete --purge %s", name)
	operation := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		HelmCmd(deleteCmd)

		err := HelmCmd(installCmd)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	notify := newNotify(name + " install")
	err = backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, c := range conditions {
		err = waitFor(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (h *Host) InstallCertResource() error {
	{
		log.Println("level", "debug", "message", "installing cert resource chart")

		o := func() error {
			// NOTE we ignore errors here because we cannot get really useful error
			// handling done. This here should anyway only be a quick fix until we use
			// the helm client lib. Then error handling will be better.
			HelmCmd("delete --purge cert-config-e2e")

			err := HelmCmd("registry install quay.io/giantswarm/apiextensions-cert-config-e2e-chart:stable -- -n cert-config-e2e --set commonDomain=${COMMON_DOMAIN} --set clusterName=${CLUSTER_NAME}")
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := NewExponentialBackoff(ShortMaxWait, ShortMaxInterval)
		n := newNotify("cert-config-e2e-chart install")
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		log.Println("level", "debug", "message", "installed cert resource chart")
	}

	{
		log.Println("level", "debug", "message", "waiting for k8s secret to be there")

		o := func() error {
			n := fmt.Sprintf("%s-api", os.Getenv("CLUSTER_NAME"))
			_, err := h.k8sClient.CoreV1().Secrets("default").Get(n, metav1.GetOptions{})
			if err != nil {
				// TODO remove this when not needed for debugging anymore
				fmt.Printf("%#v\n", err)
				return microerror.Maskf(waitError, "k8s secret is still missing")
			}

			return nil
		}
		b := NewExponentialBackoff(ShortMaxWait, ShortMaxInterval)
		n := func(err error, delay time.Duration) {
			log.Println("level", "debug", "message", err.Error())
		}

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		log.Println("level", "debug", "message", "k8s secret is there")
	}

	return nil
}

// K8sClient returns the host cluster framework's Kubernetes client.
func (h *Host) K8sClient() kubernetes.Interface {
	return h.k8sClient
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

// RestConfig returns the host cluster framework's rest config.
func (h *Host) RestConfig() *rest.Config {
	return h.restConfig
}

func (h *Host) Setup() error {
	if err := h.CreateNamespace("giantswarm"); err != nil {
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

func (h *Host) installVault() error {
	operation := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		HelmCmd("delete --purge vault")

		err := HelmCmd("registry install quay.io/giantswarm/vaultlab-chart:stable -- --set vaultToken=${VAULT_TOKEN} -n vault")
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	notify := newNotify("vaultlab-chart install")
	err := backoff.RetryNotify(operation, h.backoff, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	return waitFor(h.runningPod("default", "app=vault"))
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

func (h *Host) secret(namespace, secretName string) func() error {
	return func() error {
		_, err := h.k8sClient.CoreV1().
			Secrets(namespace).
			Get(secretName, metav1.GetOptions{})
		return microerror.Mask(err)
	}
}
