package controller

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgname"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/drainerinitializer"
	"github.com/giantswarm/aws-operator/service/controller/resource/drainfinisher"
)

type controlPlaneDrainerResourceSetConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	HostAWSConfig aws.Config
	ProjectName   string
}

func newControlPlaneDrainerResourceSet(config controlPlaneDrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var asgNameResource resource.Interface
	{
		c := asgname.Config{
			Logger: config.Logger,

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
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			CPAWSConfig:   config.HostAWSConfig,
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerInitializerResource resource.Interface
	{
		c := drainerinitializer.ResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			LabelMapFunc:  controlPlaneDrainerLabelMapFunc,
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),
		}

		drainerInitializerResource, err = drainerinitializer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainFinisherResource resource.Interface
	{
		c := drainfinisher.ResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			LabelMapFunc:      controlPlaneDrainerLabelMapFunc,
			LifeCycleHookName: key.LifeCycleHookControlPlane,
		}

		drainFinisherResource, err = drainfinisher.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		awsClientResource,
		asgNameResource,
		asgStatusResource,
		drainerInitializerResource,
		drainFinisherResource,
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

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToControlPlane(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == project.Version() {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func controlPlaneDrainerLabelMapFunc(cr metav1.Object) map[string]string {
	return map[string]string{
		label.Cluster:      key.ClusterID(cr),
		label.ControlPlane: key.ControlPlaneID(cr),
	}
}
