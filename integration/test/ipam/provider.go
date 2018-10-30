// +build k8srequired

package ipam

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/release"
	e2eclientsaws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderConfig struct {
	AWSClient *e2eclientsaws.Client
	Host      *framework.Host
	Logger    micrologger.Logger
	Release   *release.Release
}

type Provider struct {
	awsClient *e2eclientsaws.Client
	host      *framework.Host
	logger    micrologger.Logger
	release   *release.Release
}

func NewProvider(config ProviderConfig) (*Provider, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.Host == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Host must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Release == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Release must not be empty", config)
	}

	p := &Provider{
		awsClient: config.AWSClient,
		host:      config.Host,
		logger:    config.Logger,
		release:   config.Release,
	}

	return p, nil
}

func (p *Provider) CreateCluster(ctx context.Context, id string) error {
	setupConfig, err := p.newSetupConfig(id)
	if err != nil {
		return microerror.Mask(err)
	}

	err = setup.InstallAWSConfig(ctx, id, setupConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Provider) DeleteCluster(ctx context.Context, id string) error {
	setupConfig, err := p.newSetupConfig(id)
	if err != nil {
		return microerror.Mask(err)
	}

	err = p.release.EnsureDeleted(ctx, "apiextensions-aws-config-e2e", setup.CRNotExistsCondition(ctx, id, setupConfig))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Provider) GetClusterStatus(ctx context.Context, id string) (v1alpha1.StatusCluster, error) {
	customResource, err := p.host.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(id, metav1.GetOptions{})
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return customResource.ClusterStatus(), nil
}

func (p *Provider) WaitForClusterCreated(ctx context.Context, id string) error {
	setupConfig, err := p.newSetupConfig(id)
	if err != nil {
		return microerror.Mask(err)
	}

	err = setupConfig.Guest.WaitForGuestReady()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Provider) WaitForClusterDeleted(ctx context.Context, id string) error {
	setupConfig, err := p.newSetupConfig(id)
	if err != nil {
		return microerror.Mask(err)
	}

	err = setupConfig.Guest.WaitForAPIDown()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Provider) newSetupConfig(id string) (setup.Config, error) {
	var err error

	var newTenant *framework.Guest
	{
		c := framework.GuestConfig{
			Logger: p.logger,

			ClusterID:    id,
			CommonDomain: env.CommonDomain(),
		}

		newTenant, err = framework.NewGuest(c)
		if err != nil {
			return setup.Config{}, microerror.Mask(err)
		}
	}

	err = newTenant.Initialize()
	if err != nil {
		return setup.Config{}, microerror.Mask(err)
	}

	setupConfig := setup.Config{
		AWSClient: p.awsClient,
		Guest:     newTenant,
		Host:      p.host,
		Logger:    p.logger,
		Release:   p.release,
	}

	return setupConfig, nil
}
