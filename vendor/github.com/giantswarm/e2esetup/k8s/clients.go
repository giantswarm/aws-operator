package k8s

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/e2esetup/chart/env"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientsConfig is DEPRECATED. Use https://github.com/giantswarm/k8sclient/
// instead.
type ClientsConfig struct {
	Logger micrologger.Logger

	KubeConfigPath string
}

type Clients struct {
	logger micrologger.Logger

	extClient  *apiextensionsclient.Clientset
	g8sClient  *versioned.Clientset
	k8sClient  *kubernetes.Clientset
	restConfig *rest.Config
}

// NewClients is DEPRECATED. Use https://github.com/giantswarm/k8sclient/
// instead.
func NewClients(config ClientsConfig) (*Clients, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.KubeConfigPath == "" {
		config.KubeConfigPath = env.KubeConfigPath()
	}

	var err error

	restConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeConfigPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var extClient *apiextensionsclient.Clientset
	{
		c := rest.CopyConfig(restConfig)

		extClient, err = apiextensionsclient.NewForConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var g8sClient *versioned.Clientset
	{
		c := rest.CopyConfig(restConfig)

		g8sClient, err = versioned.NewForConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient *kubernetes.Clientset
	{
		c := rest.CopyConfig(restConfig)

		k8sClient, err = kubernetes.NewForConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Clients{
		logger: config.Logger,

		extClient:  extClient,
		g8sClient:  g8sClient,
		k8sClient:  k8sClient,
		restConfig: restConfig,
	}

	return c, nil
}

func (c *Clients) ExtClient() apiextensionsclient.Interface {
	return c.extClient
}

func (c *Clients) G8sClient() versioned.Interface {
	return c.g8sClient
}

func (c *Clients) K8sClient() kubernetes.Interface {
	return c.k8sClient
}

func (c *Clients) RestConfig() *rest.Config {
	return rest.CopyConfig(c.restConfig)
}
