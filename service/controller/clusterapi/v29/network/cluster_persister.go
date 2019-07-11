package network

import (
	"context"
	"encoding/json"
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

type ClusterPersisterConfig struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

type ClusterPersister struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func NewClusterPersister(config ClusterPersisterConfig) (*ClusterPersister, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	p := &ClusterPersister{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return p, nil
}

func (p *ClusterPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	cr, err := p.cmaClient.ClusterV1alpha1().Clusters(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	var providerStatus v1alpha1.AWSClusterStatus
	{
		err := json.Unmarshal(cr.Status.ProviderStatus.Raw, &providerStatus)
		if err != nil {
			return microerror.Mask(err)
		}

		providerStatus.Provider.Network.CIDR = subnet.String()
	}

	{
		b, err := json.Marshal(providerStatus)
		if err != nil {
			return microerror.Mask(err)
		}

		cr.Status.ProviderStatus.Raw = b
	}

	{
		_, err := p.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).UpdateStatus(cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
