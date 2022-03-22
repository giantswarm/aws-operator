package ipam

import (
	"context"
	"net"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/annotation"
)

type MachineDeploymentPersisterConfig struct {
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger
}

type MachineDeploymentPersister struct {
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger
}

func NewMachineDeploymentPersister(config MachineDeploymentPersisterConfig) (*MachineDeploymentPersister, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	p := &MachineDeploymentPersister{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return p, nil
}

func (p *MachineDeploymentPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	var cr *infrastructurev1alpha3.AWSMachineDeployment
	err := p.ctrlClient.Get(ctx, ctrlClient.ObjectKey{Name: name, Namespace: namespace}, cr)
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
		err = p.ctrlClient.Update(ctx, cr, &ctrlClient.UpdateOptions{Raw: &metav1.UpdateOptions{}})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
