package framework

import (
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/e2e-harness/pkg/harness"
)

const (
	// minimumNodesReady represents the minimun number of ready nodes in a guest
	// cluster to be considered healthy.
	minimumNodesReady = 3
)

type Guest struct {
	k8sClient  kubernetes.Interface
	restConfig *rest.Config
}

func NewGuest() (*Guest, error) {
	g := &Guest{
		k8sClient:  nil,
		restConfig: nil,
	}

	return g, nil
}

// K8sClient returns the guest cluster framework's Kubernetes client. The client
// being returned is properly configured ones Guest.Setup() got executed
// successfully.
func (g *Guest) K8sClient() kubernetes.Interface {
	return g.k8sClient
}

// RestConfig returns the guest cluster framework's rest config. The config
// being returned is properly configured ones Guest.Setup() got executed
// successfully.
func (g *Guest) RestConfig() *rest.Config {
	return g.restConfig
}

// Setup provides a separate initialization step because of the nature of the
// host/guest cluster design. We have to setup things in different stages.
// Constructing the frameworks can be done right away but setting them up can
// only happen as soon as certain requirements have been met. A requirement for
// the guest framework is a set up host cluster.
func (g *Guest) Setup() error {
	var err error

	var hostK8sClient kubernetes.Interface
	{
		c, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
		if err != nil {
			return microerror.Mask(err)
		}
		hostK8sClient, err = kubernetes.NewForConfig(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var guestK8sClient kubernetes.Interface
	var guestRestConfig *rest.Config
	{
		n := os.ExpandEnv("${CLUSTER_NAME}-api")
		s, err := hostK8sClient.CoreV1().Secrets("default").Get(n, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		guestRestConfig = &rest.Config{
			Host: os.ExpandEnv("https://api.${CLUSTER_NAME}.${COMMON_DOMAIN_GUEST}"),
			TLSClientConfig: rest.TLSClientConfig{
				CAData:   s.Data["ca"],
				CertData: s.Data["crt"],
				KeyData:  s.Data["key"],
			},
		}

		guestK8sClient, err = kubernetes.NewForConfig(guestRestConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	g.k8sClient = guestK8sClient
	g.restConfig = guestRestConfig

	err = g.WaitForGuestReady()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (g *Guest) WaitForAPIDown() error {
	time.Sleep(1 * time.Second)

	log.Println("level", "debug", "message", "waiting for k8s API to be down")

	o := func() error {
		_, err := g.k8sClient.CoreV1().Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return nil
		}

		return microerror.Maskf(waitError, "k8s API is still up")
	}
	b := newExponentialBackoff(ShortMaxWait, ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Println("level", "debug", "message", "k8s API is down")

	return nil
}

func (g *Guest) WaitForAPIUp() error {
	log.Println("level", "debug", "message", "waiting for k8s API to be up")

	o := func() error {
		_, err := g.k8sClient.CoreV1().Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return microerror.Maskf(waitError, "k8s API is still down")
		}

		return nil
	}
	b := newExponentialBackoff(LongMaxWait, LongMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Println("level", "debug", "message", "k8s API is up")

	return nil
}

func (g *Guest) WaitForGuestReady() error {
	var err error

	err = g.WaitForAPIUp()
	if err != nil {
		return microerror.Mask(err)
	}

	err = g.WaitForNodesUp(minimumNodesReady)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (g *Guest) WaitForNodesUp(numberOfNodes int) error {
	log.Println("level", "debug", "message", "waiting for k8s nodes to be up")

	o := func() error {
		nodes, err := g.k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(nodes.Items) != numberOfNodes {
			return microerror.Maskf(waitError, "worker nodes are still not found")
		}

		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == v1.NodeReady && c.Status != v1.ConditionTrue {
					return microerror.Maskf(waitError, "worker nodes are still not ready")
				}
			}
		}

		return nil
	}
	b := newExponentialBackoff(LongMaxWait, LongMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Println("level", "debug", "message", "k8s nodes are up")

	return nil
}
