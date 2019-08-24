// +build k8srequired

package scaling

import (
	"context"
	"encoding/json"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ProviderConfig struct {
	G8sClient      versioned.Interface
	GuestFramework *framework.Guest
	Logger         micrologger.Logger

	ClusterID string
}

type Provider struct {
	g8sClient      versioned.Interface
	guestFramework *framework.Guest
	logger         micrologger.Logger

	clusterID string
}

func NewProvider(config ProviderConfig) (*Provider, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	p := &Provider{
		guestFramework: config.GuestFramework,
		g8sClient:      config.G8sClient,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return p, nil
}

func (p *Provider) AddWorker() error {
	// TODO remove the legacy approach when v22 resources are gone in the
	// aws-operator.
	{
		customObject, err := p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		patches := []Patch{
			{
				Op:    "add",
				Path:  "/spec/aws/workers/-",
				Value: customObject.Spec.AWS.Workers[0],
			},
		}

		b, err := json.Marshal(patches)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Patch(p.clusterID, types.JSONPatchType, b)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		customObject, err := p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Spec.Cluster.Scaling.Max++
		customObject.Spec.Cluster.Scaling.Min++

		_, err = p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Update(customObject)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (p *Provider) NumMasters() (int, error) {
	customObject, err := p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.AWS.Masters)

	return num, nil
}

func (p *Provider) NumWorkers() (int, error) {
	customObject, err := p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.AWS.Workers)

	return num, nil
}

func (p *Provider) RemoveWorker() error {
	// TODO remove the legacy approach when v22 resources are gone in the
	// aws-operator.
	{
		patches := []Patch{
			{
				Op:   "remove",
				Path: "/spec/aws/workers/1",
			},
		}

		b, err := json.Marshal(patches)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Patch(p.clusterID, types.JSONPatchType, b)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		customObject, err := p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Spec.Cluster.Scaling.Max--
		customObject.Spec.Cluster.Scaling.Min--

		_, err = p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Update(customObject)
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
