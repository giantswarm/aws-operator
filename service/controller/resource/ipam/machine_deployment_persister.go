package ipam

import (
	"context"
	"net"

	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	cr, err := p.g8sClient.InfrastructureV1alpha2().AWSMachineDeployments(namespace).Get(ctx, name, metav1.GetOptions{})
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
		_, err := p.g8sClient.InfrastructureV1alpha2().AWSMachineDeployments(namespace).Update(ctx, cr, metav1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
