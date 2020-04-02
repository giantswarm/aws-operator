package controller

import (
	"context"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/resource/drainer"
	"github.com/giantswarm/aws-operator/service/controller/resource/drainfinisher"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/internal/credential"
)

type drainerResourceSetConfig struct {
	ControlPlaneAWSClients aws.Clients
	HostAWSConfig          aws.Config
	K8sClient              k8sclient.Interface
	Logger                 micrologger.Logger

	ProjectName    string
	Route53Enabled bool
}

func newDrainerResourceSet(config drainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var asgStatusResource resource.Interface
	{
		c := asgstatus.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		asgStatusResource, err = asgstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerResource resource.Interface
	{
		c := drainer.ResourceConfig{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		drainerResource, err = drainer.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainFinisherResource resource.Interface
	{
		c := drainfinisher.ResourceConfig{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		drainFinisherResource, err = drainfinisher.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpOutputsResource resource.Interface
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

	resources := []resource.Interface{
		tccpOutputsResource,
		asgStatusResource,
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

		if key.VersionBundleVersion(customObject) == project.Version() {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		var tenantClusterAWSClients aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient.K8sClient(), obj)
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
