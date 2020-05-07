package controller

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
)

type MachineDeploymentDrainerConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	HostAWSConfig aws.Config
}

type MachineDeploymentDrainerConfigLabelSelector struct {
	Enabled          bool
	OverridenVersion string
}

type MachineDeploymentDrainer struct {
	*controller.Controller
}

func NewMachineDeploymentDrainer(config MachineDeploymentDrainerConfig) (*MachineDeploymentDrainer, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resourceSets, err := newMachineDeploymentDrainerResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-drainer-controller.
			Name: project.Name() + "-drainer-controller",
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSMachineDeployment)
			},
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &MachineDeploymentDrainer{
		Controller: operatorkitController,
	}

	return d, nil
}

func newMachineDeploymentDrainerResourceSets(config MachineDeploymentDrainerConfig) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := machineDeploymentDrainerResourceSetConfig(config)

		resourceSet, err = newMachineDeploymentDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
