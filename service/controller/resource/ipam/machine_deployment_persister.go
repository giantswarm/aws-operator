package ipam

import (
	"context"
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/pkg/annotation"
)

type MachineDeploymentPersisterConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type MachineDeploymentPersister struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func NewMachineDeploymentPersister(config MachineDeploymentPersisterConfig) (*MachineDeploymentPersister, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	p := &MachineDeploymentPersister{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return p, nil
}

func (p *MachineDeploymentPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	n := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	cr, err := p.g8sClient.InfrastructureV1alpha2().AWSMachineDeployments().Get(n.String(), metav1.GetOptions{})
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
		_, err := p.g8sClient.InfrastructureV1alpha2().AWSMachineDeployments().Update(cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
