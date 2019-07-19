package ipam

import (
	"context"
	"net"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/pkg/annotation"
)

type MachineDeploymentPersisterConfig struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

type MachineDeploymentPersister struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func NewMachineDeploymentPersister(config MachineDeploymentPersisterConfig) (*MachineDeploymentPersister, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	p := &MachineDeploymentPersister{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return p, nil
}

func (p *MachineDeploymentPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	cr, err := p.cmaClient.ClusterV1alpha1().MachineDeployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	{
		if cr.Annotations == nil {
			cr.Annotations = map[string]string{}
		}

		cr.Annotations[annotation.MachineDeploymentSubnet] = subnet.String()
	}

	{
		_, err := p.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).Update(cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
