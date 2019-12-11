package ipam

import (
	"context"
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterPersisterConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type ClusterPersister struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func NewClusterPersister(config ClusterPersisterConfig) (*ClusterPersister, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	p := &ClusterPersister{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return p, nil
}

func (p *ClusterPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	cr, err := p.g8sClient.InfrastructureV1alpha2().AWSClusters(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	cr.Status.Provider.Network.CIDR = subnet.String()

	_, err = p.g8sClient.InfrastructureV1alpha2().AWSClusters(namespace).UpdateStatus(cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
