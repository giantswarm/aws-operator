// +build k8srequired

package ipam

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	e2eclientsaws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/integration/setup"
)

type ProviderConfig struct {
	AWSClient  *e2eclientsaws.Client
	K8sClients *k8sclient.Clients
	Logger     micrologger.Logger
	Release    *release.Release
}

type Provider struct {
	awsClient  *e2eclientsaws.Client
	k8sClients *k8sclient.Clients
	logger     micrologger.Logger
	release    *release.Release
}

func NewProvider(config ProviderConfig) (*Provider, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.K8sClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Release == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Release must not be empty", config)
	}

	p := &Provider{
		awsClient:  config.AWSClient,
		k8sClients: config.K8sClients,
		logger:     config.Logger,
		release:    config.Release,
	}

	return p, nil
}

func (p *Provider) CreateCluster(ctx context.Context, id string) error {
	setupConfig := setup.Config{
		AWSClient:  p.awsClient,
		K8sClients: p.k8sClients,
		Logger:     p.logger,
		Release:    p.release,
	}

	wait := false
	err := setup.EnsureTenantClusterCreated(ctx, id, setupConfig, wait)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Provider) DeleteCluster(ctx context.Context, id string) error {
	setupConfig := setup.Config{
		AWSClient:  p.awsClient,
		K8sClients: p.k8sClients,
		Logger:     p.logger,
		Release:    p.release,
	}

	wait := false
	err := setup.EnsureTenantClusterDeleted(ctx, id, setupConfig, wait)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Provider) GetClusterStatus(ctx context.Context, id string) (v1alpha1.StatusCluster, error) {
	customResource, err := p.k8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(id, metav1.GetOptions{})
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return customResource.ClusterStatus(), nil
}

func (p *Provider) WaitForClusterCreated(ctx context.Context, id string) error {
	p.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for tenant cluster %#q to be created", id))

	o := func() error {
		customResource, err := p.k8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(id, metav1.GetOptions{})
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
		_, err := p.k8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(id, metav1.GetOptions{})
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
