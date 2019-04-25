package v26

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/credential"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/key"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/resource/drainer"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/resource/drainfinisher"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26/resource/workerasgname"
)

type DrainerResourceSetConfig struct {
	ControlPlaneAWSClients aws.Clients
	G8sClient              versioned.Interface
	HostAWSConfig          aws.Config
	K8sClient              kubernetes.Interface
	Logger                 micrologger.Logger

	ProjectName    string
	Route53Enabled bool
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var drainerResource controller.Resource
	{
		c := drainer.ResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		drainerResource, err = drainer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainFinisherResource controller.Resource
	{
		c := drainfinisher.ResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		drainFinisherResource, err = drainfinisher.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpOutputsResource controller.Resource
	{
		c := tccpoutputs.Config{
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		tccpOutputsResource, err = tccpoutputs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerASGNameResource controller.Resource
	{
		c := workerasgname.ResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		workerASGNameResource, err = workerasgname.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		tccpOutputsResource,
		workerASGNameResource,
		drainerResource,
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
		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(customObject) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		var tenantClusterAWSClients aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient, obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			c := config.HostAWSConfig
			c.RoleARN = arn

			tenantClusterAWSClients, err = aws.NewClients(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		cc := controllercontext.Context{
			Client: controllercontext.ContextClient{
				ControlPlane: controllercontext.ContextClientControlPlane{
					AWS: config.ControlPlaneAWSClients,
				},
				TenantCluster: controllercontext.ContextClientTenantCluster{
					AWS: tenantClusterAWSClients,
				},
			},
		}
		ctx = controllercontext.NewContext(ctx, cc)

		return ctx, nil
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
