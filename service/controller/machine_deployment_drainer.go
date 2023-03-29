package controller

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v8/pkg/controller"
	"github.com/giantswarm/operatorkit/v8/pkg/resource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/retryresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v14/client/aws"
	"github.com/giantswarm/aws-operator/v14/pkg/label"
	"github.com/giantswarm/aws-operator/v14/pkg/project"
	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
	"github.com/giantswarm/aws-operator/v14/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/v14/service/controller/resource/drainerfinalizer"
	"github.com/giantswarm/aws-operator/v14/service/controller/resource/drainerinitializer"
	"github.com/giantswarm/aws-operator/v14/service/internal/asg"
	event "github.com/giantswarm/aws-operator/v14/service/internal/recorder"
)

type MachineDeploymentDrainerConfig struct {
	Event     event.Interface
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	HostAWSConfig aws.Config
}

type MachineDeploymentDrainer struct {
	*controller.Controller
}

func NewMachineDeploymentDrainer(config MachineDeploymentDrainerConfig) (*MachineDeploymentDrainer, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	var resources []resource.Interface
	{
		resources, err = newMachineDeploymentDrainerResources(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			InitCtx: func(ctx context.Context, obj interface{}) (context.Context, error) {
				return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() ctrlClient.Object {
				return new(infrastructurev1alpha3.AWSMachineDeployment)
			},
			Resources:    resources,
			ResyncPeriod: key.DrainerResyncPeriod,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-drainer-controller.
			Name: project.Name() + "-drainer-controller",
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

func newMachineDeploymentDrainerResources(config MachineDeploymentDrainerConfig) ([]resource.Interface, error) {
	var err error

	var newASG asg.Interface
	{
		c := asg.Config{
			Stack:        key.StackTCNP,
			TagKey:       key.TagMachineDeployment,
			TagValueFunc: key.MachineDeploymentID,
		}

		newASG, err = asg.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsClientResource resource.Interface
	{
		c := awsclient.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			CPAWSConfig:   config.HostAWSConfig,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerInitializerResource resource.Interface
	{
		c := drainerinitializer.ResourceConfig{
			ASG:        newASG,
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,

			LabelMapFunc:      machineDeploymentDrainerLabelMapFunc,
			LifeCycleHookName: key.LifeCycleHookNodePool,
			ToClusterFunc:     newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		drainerInitializerResource, err = drainerinitializer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerFinalizerResource resource.Interface
	{
		c := drainerfinalizer.ResourceConfig{
			ASG:        newASG,
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,

			LabelMapFunc:      machineDeploymentDrainerLabelMapFunc,
			LifeCycleHookName: key.LifeCycleHookNodePool,
		}

		drainerFinalizerResource, err = drainerfinalizer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		awsClientResource,
		drainerInitializerResource,
		drainerFinalizerResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}

func machineDeploymentDrainerLabelMapFunc(cr metav1.Object) map[string]string {
	return map[string]string{
		label.Cluster:           key.ClusterID(cr),
		label.MachineDeployment: key.MachineDeploymentID(cr),
		label.OperatorVersion:   key.OperatorVersion(cr),
	}
}
