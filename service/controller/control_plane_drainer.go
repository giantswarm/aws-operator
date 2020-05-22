package controller

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgname"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/drainerfinalizer"
	"github.com/giantswarm/aws-operator/service/controller/resource/drainerinitializer"
)

type ControlPlaneDrainerConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	HostAWSConfig aws.Config
}

type ControlPlaneDrainer struct {
	*controller.Controller
}

func NewControlPlaneDrainer(config ControlPlaneDrainerConfig) (*ControlPlaneDrainer, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	var resources []resource.Interface
	{
		resources, err = newControlPlaneDrainerResources(config)
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
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSControlPlane)
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

	d := &ControlPlaneDrainer{
		Controller: operatorkitController,
	}

	return d, nil
}

func newControlPlaneDrainerResources(config ControlPlaneDrainerConfig) ([]resource.Interface, error) {
	var err error

	var asgNameResource resource.Interface
	{
		c := asgname.Config{
			Logger: config.Logger,

			Stack:        key.StackTCCPN,
			TagKey:       key.TagControlPlane,
			TagValueFunc: key.ControlPlaneID,
		}

		asgNameResource, err = asgname.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var asgStatusResource resource.Interface
	{
		c := asgstatus.Config{
			Logger: config.Logger,
		}

		asgStatusResource, err = asgstatus.New(c)
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerInitializerResource resource.Interface
	{
		c := drainerinitializer.ResourceConfig{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,

			LabelMapFunc:      controlPlaneDrainerLabelMapFunc,
			ToClusterFunc:     newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
			LifeCycleHookName: key.LifeCycleHookControlPlane,
		}

		drainerInitializerResource, err = drainerinitializer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerFinalizerResource resource.Interface
	{
		c := drainerfinalizer.ResourceConfig{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,

			LabelMapFunc:      controlPlaneDrainerLabelMapFunc,
			LifeCycleHookName: key.LifeCycleHookControlPlane,
		}

		drainerFinalizerResource, err = drainerfinalizer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		awsClientResource,
		asgNameResource,
		asgStatusResource,
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

func controlPlaneDrainerLabelMapFunc(cr metav1.Object) map[string]string {
	return map[string]string{
		label.Cluster:         key.ClusterID(cr),
		label.ControlPlane:    key.ControlPlaneID(cr),
		label.OperatorVersion: key.OperatorVersion(cr),
	}
}
