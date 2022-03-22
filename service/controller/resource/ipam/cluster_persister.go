package ipam

import (
	"context"
	"net"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterPersisterConfig struct {
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger
}

type ClusterPersister struct {
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger
}

func NewClusterPersister(config ClusterPersisterConfig) (*ClusterPersister, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	p := &ClusterPersister{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return p, nil
}

func (p *ClusterPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	cr := &infrastructurev1alpha3.AWSCluster{}
	err := p.ctrlClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	cr.Status.Provider.Network.CIDR = subnet.String()

	err = p.ctrlClient.Status().Update(ctx, cr, &ctrlClient.UpdateOptions{Raw: &metav1.UpdateOptions{}})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
