package setup

import (
	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/filelogger"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	"github.com/giantswarm/e2e-harness/pkg/release"
	e2eclientsaws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	namespace       = "default"
	organization    = "giantswarm"
	tillerNamespace = "kube-system"
	quayAddress     = "https://quay.io"
)

type Config struct {
	AWSClient *e2eclientsaws.Client
	Guest     *framework.Guest
	Host      *framework.Host
	Release   *release.Release
	Resource  *resource.Resource
	Logger    micrologger.Logger
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

	var guest *framework.Guest
	{
		c := framework.GuestConfig{
			Logger: logger,

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

			ClusterID:  env.ClusterID(),
			VaultToken: env.VaultToken(),
		}

		host, err = framework.NewHost(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var fileLogger *filelogger.FileLogger
	{
		c := filelogger.Config{
			Backoff:   backoff.NewExponential(backoff.ShortMaxWait, backoff.LongMaxInterval),
			K8sClient: host.K8sClient(),
			Logger:    logger,
		}

		fileLogger, err = filelogger.New(c)
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
			FileLogger: fileLogger,
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

	var newResource *resource.Resource
	{
		c := resource.Config{
			HelmClient: helmClient,
			Logger:     logger,

			Namespace: namespace,
		}

		newResource, err = resource.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}

	}

	c := Config{
		AWSClient: awsClient,
		Guest:     guest,
		Host:      host,
		Release:   newRelease,
		Resource:  newResource,
		Logger:    logger,
	}

	return c, nil
}
