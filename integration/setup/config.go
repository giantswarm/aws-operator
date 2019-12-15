// +build k8srequired

package setup

import (
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/release"
	e2eclientsaws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/integration/env"
)

const (
	namespace       = "giantswarm"
	tillerNamespace = "kube-system"
)

type Config struct {
	AWSClient  *e2eclientsaws.Client
	Guest      *framework.Guest
	HelmClient helmclient.Interface
	Host       *framework.Host
	K8s        *k8sclient.Setup
	Logger     micrologger.Logger
	Release    *release.Release

	// UseDefaultTenant defines whether the standard test setup should ensure the
	// default tenant cluster. This is enabled by default. Most tests simply use
	// the standard test tenant cluster. One exception is IPAM. There we launch
	// multiple tenant clusters and do not want to make use of the default one.
	UseDefaultTenant bool
}

func NewConfig() (Config, error) {
	var err error

	var awsClient *e2eclientsaws.Client
	{
		awsClient, err = e2eclientsaws.NewClient()
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var cpK8sClients *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger: logger,
			SchemeBuilder: k8sclient.SchemeBuilder{
				corev1alpha1.AddToScheme,
				providerv1alpha1.AddToScheme,
			},

			KubeConfigPath: harness.DefaultKubeConfig,
		}

		cpK8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var guest *framework.Guest
	{
		c := framework.GuestConfig{
			HostK8sClient: cpK8sClients.K8sClient(),
			Logger:        logger,

			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}

		guest, err = framework.NewGuest(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var host *framework.Host
	{
		c := framework.HostConfig{
			Logger: logger,

			ClusterID: env.ClusterID(),
		}

		host, err = framework.NewHost(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var k8sSetup *k8sclient.Setup
	{
		c := k8sclient.SetupConfig{
			Clients: cpK8sClients,
			Logger:  logger,
		}

		k8sSetup, err = k8sclient.NewSetup(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:    logger,
			K8sClient: host.K8sClient(),

			RestConfig:      host.RestConfig(),
			TillerNamespace: tillerNamespace,
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var newRelease *release.Release
	{
		c := release.Config{
			ExtClient:  host.ExtClient(),
			G8sClient:  host.G8sClient(),
			HelmClient: helmClient,
			K8sClient:  host.K8sClient(),
			Logger:     logger,

			Namespace: namespace,
		}

		newRelease, err = release.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	c := Config{
		AWSClient:  awsClient,
		Guest:      guest,
		HelmClient: helmClient,
		Host:       host,
		K8s:        k8sSetup,
		Logger:     logger,
		Release:    newRelease,

		UseDefaultTenant: true,
	}

	return c, nil
}
