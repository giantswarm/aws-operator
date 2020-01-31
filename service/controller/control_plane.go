package controller

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
)

type ControlPlaneConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	HostAWSConfig  aws.Config
	Route53Enabled bool
}

type ControlPlane struct {
	*controller.Controller
}

func NewControlPlane(config ControlPlaneConfig) (*ControlPlane, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resourceSets, err := newControlPlaneResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          infrastructurev1alpha3.NewAWSControlPlaneCRD(),
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-drainer-controller.
			Name: project.Name() + "-control-plane-controller",
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSMachineDeployment)
			},
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &ControlPlane{
		Controller: operatorkitController,
	}

	return d, nil
}

func newControlPlaneResourceSets(config ControlPlaneConfig) ([]*controller.ResourceSet, error) {
	var err error

	var controlPlaneAWSClients aws.Clients
	{
		c := aws.Config{
			AccessKeyID:     config.HostAWSConfig.AccessKeyID,
			AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
			Region:          config.HostAWSConfig.Region,
			SessionToken:    config.HostAWSConfig.SessionToken,
		}

		controlPlaneAWSClients, err = aws.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSet *controller.ResourceSet
	{
		c := drainerResourceSetConfig{
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.K8sClient.G8sClient(),
			K8sClient:              config.K8sClient.K8sClient(),
			Logger:                 config.Logger,

			HostAWSConfig:  config.HostAWSConfig,
			Route53Enabled: config.Route53Enabled,
		}

		resourceSet, err = newControlPlaneResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
