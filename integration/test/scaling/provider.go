// +build k8srequired

package scaling

import (
	"context"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderConfig struct {
	GuestFramework *framework.Guest
	HostFramework  *framework.Host
	Logger         micrologger.Logger

	ClusterID string
}

type Provider struct {
	guestFramework *framework.Guest
	hostFramework  *framework.Host
	logger         micrologger.Logger

	clusterID string
}

func NewProvider(config ProviderConfig) (*Provider, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	p := &Provider{
		guestFramework: config.GuestFramework,
		hostFramework:  config.HostFramework,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return p, nil
}

func (p *Provider) AddWorker() error {
	{
		customObject, err := p.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Spec.Cluster.Scaling.Max++
		customObject.Spec.Cluster.Scaling.Min++

		_, err = p.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Update(customObject)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (p *Provider) NumMasters() (int, error) {
	customObject, err := p.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.AWS.Masters)

	return num, nil
}

func (p *Provider) NumWorkers() (int, error) {
	customObject, err := p.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.AWS.Workers)

	return num, nil
}

func (p *Provider) RemoveWorker() error {
	{
		customObject, err := p.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Spec.Cluster.Scaling.Max--
		customObject.Spec.Cluster.Scaling.Min--

		_, err = p.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Update(customObject)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (p *Provider) WaitForNodes(ctx context.Context, num int) error {
	err := p.guestFramework.WaitForNodesReady(ctx, num)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
