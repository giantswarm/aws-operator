package ipam

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/key"

	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type MachineDeploymentCheckerConfig struct {
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger
}

type MachineDeploymentChecker struct {
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger
}

func NewMachineDeploymentChecker(config MachineDeploymentCheckerConfig) (*MachineDeploymentChecker, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &MachineDeploymentChecker{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return a, nil
}

func (c *MachineDeploymentChecker) Check(ctx context.Context, namespace string, name string) (bool, error) {

	objectKey := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	cr := &infrastructurev1alpha3.AWSMachineDeployment{}
	err := c.ctrlClient.Get(ctx, objectKey, cr)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// We check the subnet we want to ensure in the CR annotations. In case there
	// is no subnet tracked so far, we want to proceed with the allocation
	// process. Thus we return true.
	if key.MachineDeploymentSubnet(*cr) == "" {
		return true, nil
	}

	// At this point the subnet is already allocated for the CR we check here. So
	// we do not want to proceed further and return false.
	return false, nil
}
