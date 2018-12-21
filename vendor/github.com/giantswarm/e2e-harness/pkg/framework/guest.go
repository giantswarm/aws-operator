package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
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

type GuestConfig struct {
	Logger micrologger.Logger

	ClusterID    string
	CommonDomain string
}

type Guest struct {
	logger micrologger.Logger

	g8sClient  versioned.Interface
	k8sClient  kubernetes.Interface
	restConfig *rest.Config

	clusterID    string
	commonDomain string
}

func NewGuest(config GuestConfig) (*Guest, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.CommonDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}

	g := &Guest{
		logger: config.Logger,

		g8sClient:  nil,
		k8sClient:  nil,
		restConfig: nil,

		clusterID:    config.ClusterID,
		commonDomain: config.CommonDomain,
	}

	return g, nil
}

// G8sClient returns the guest cluster framework's apiextensions clientset. The
// client being returned is properly configured once Guest.Setup() is executed
// successfully.
func (g *Guest) G8sClient() versioned.Interface {
	return g.g8sClient
}

// K8sClient returns the guest cluster framework's Kubernetes client. The client
// being returned is properly configured once Guest.Setup() is executed
// successfully.
func (g *Guest) K8sClient() kubernetes.Interface {
	return g.k8sClient
}

// RestConfig returns the guest cluster framework's rest config. The config
// being returned is properly configured once Guest.Setup() is executed
// successfully.
func (g *Guest) RestConfig() *rest.Config {
	return g.restConfig
}

// Initialize sets up the Guest fields that are not directly injected.
func (g *Guest) Initialize() error {
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

	var guestG8sClient versioned.Interface
	var guestK8sClient kubernetes.Interface
	var guestRestConfig *rest.Config
	{
		n := fmt.Sprintf("%s-api", g.clusterID)
		s, err := hostK8sClient.CoreV1().Secrets("default").Get(n, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		guestRestConfig = &rest.Config{
			Host: fmt.Sprintf("https://api.%s.k8s.%s", g.clusterID, g.commonDomain),
			TLSClientConfig: rest.TLSClientConfig{
				CAData:   s.Data["ca"],
				CertData: s.Data["crt"],
				KeyData:  s.Data["key"],
			},
		}

		guestG8sClient, err = versioned.NewForConfig(guestRestConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		guestK8sClient, err = kubernetes.NewForConfig(guestRestConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	g.g8sClient = guestG8sClient
	g.k8sClient = guestK8sClient
	g.restConfig = guestRestConfig

	return nil
}

// Setup provides a separate initialization step because of the nature of the
// host/guest cluster design. We have to setup things in different stages.
// Constructing the frameworks can be done right away but setting them up can
// only happen as soon as certain requirements have been met. A requirement for
// the guest framework is a set up host cluster.
func (g *Guest) Setup(ctx context.Context) error {
	err := g.Initialize()
	if err != nil {
		return microerror.Mask(err)
	}

	err = g.WaitForGuestReady(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (g *Guest) WaitForAPIDown() error {
	time.Sleep(1 * time.Second)

	g.logger.Log("level", "debug", "message", "waiting for k8s API to be down")

	o := func() error {
		_, err := g.k8sClient.CoreV1().Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return nil
		}

		return microerror.Maskf(waitError, "k8s API is still up")
	}
	b := backoff.NewConstant(backoff.LongMaxWait, backoff.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		g.logger.Log("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	g.logger.Log("level", "debug", "message", "k8s API is down")

	return nil
}

func (g *Guest) WaitForAPIUp() error {
	g.logger.Log("level", "debug", "message", "waiting for k8s API to be up")

	o := func() error {
		_, err := g.k8sClient.CoreV1().Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return microerror.Maskf(waitError, "k8s API is still down")
		}

		return nil
	}
	b := backoff.NewConstant(backoff.LongMaxWait, backoff.LongMaxInterval)
	n := func(err error, delay time.Duration) {
		g.logger.Log("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	g.logger.Log("level", "debug", "message", "k8s API is up")

	return nil
}

func (g *Guest) WaitForGuestReady(ctx context.Context) error {
	var err error

	err = g.WaitForAPIUp()
	if err != nil {
		return microerror.Mask(err)
	}

	err = g.WaitForNodesReady(ctx, minimumNodesReady)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (g *Guest) WaitForNodesReady(ctx context.Context, expectedNodes int) error {
	g.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for %d k8s nodes to be in %#q state", expectedNodes, v1.NodeReady))

	o := func() error {
		nodes, err := g.k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		var nodesReady int
		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == v1.NodeReady && c.Status == v1.ConditionTrue {
					nodesReady++
				}
			}
		}

		if nodesReady != expectedNodes {
			return microerror.Maskf(waitError, "found %d/%d k8s nodes in %#q state but %d are expected", nodesReady, len(nodes.Items), v1.NodeReady, expectedNodes)
		}

		return nil
	}
	b := backoff.NewConstant(backoff.LongMaxWait, backoff.LongMaxInterval)
	n := backoff.NewNotifier(g.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	g.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for %d k8s nodes to be in %#q state", expectedNodes, v1.NodeReady))
	return nil
}
