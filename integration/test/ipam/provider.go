// +build k8srequired

package ipam

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/integration/setup"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/release"
	e2eclientsaws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/api/errors"
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
	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating tenant cluster %#q", id))

	setupConfig := setup.Config{
		AWSClient: p.awsClient,
		Host:      p.host,
		Logger:    p.logger,
		Release:   p.release,
	}

	err := setup.InstallAWSConfig(ctx, id, setupConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created tenant cluster %#q", id))

	return nil
}

func (p *Provider) DeleteCluster(ctx context.Context, id string) error {
	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting tenant cluster %#q", id))

	setupConfig := setup.Config{
		AWSClient: p.awsClient,
		Host:      p.host,
		Logger:    p.logger,
		Release:   p.release,
	}

	err := p.release.EnsureDeleted(ctx, id, setup.CRNotExistsCondition(ctx, id, setupConfig))
	if err != nil {
		return microerror.Mask(err)
	}

	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted tenant cluster %#q", id))

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
	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for tenant cluster %#q to be created", id))

	o := func() error {
		customResource, err := p.host.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(id, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		// In case the CR status indicates the tenant cluster has a Created status
		// condition, we return nil and stop retrying.
		if customResource.ClusterStatus().HasCreatedCondition() {
			return nil
		}

		return microerror.Mask(missingCreatedConditionError)
	}
	b := backoff.NewConstant(backoff.LongMaxWait, backoff.LongMaxInterval)
	n := backoff.NewNotifier(p.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for tenant cluster %#q to be created", id))

	return nil
}

func (p *Provider) WaitForClusterDeleted(ctx context.Context, id string) error {
	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for tenant cluster %#q to be deleted", id))

	o := func() error {
		_, err := p.host.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(id, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		return microerror.Mask(clusterCRStillExistsError)
	}
	b := backoff.NewConstant(backoff.LongMaxWait, backoff.LongMaxInterval)
	n := backoff.NewNotifier(p.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for tenant cluster %#q to be deleted", id))

	return nil
}
