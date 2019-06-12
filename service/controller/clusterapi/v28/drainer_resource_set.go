package v28

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/credential"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/resource/drainer"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/resource/drainfinisher"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/resource/machinedeployment"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/resource/tccpoutputs"
)

type DrainerResourceSetConfig struct {
	CMAClient              clientset.Interface
	ControlPlaneAWSClients aws.Clients
	G8sClient              versioned.Interface
	K8sClient              kubernetes.Interface
	Logger                 micrologger.Logger

	HostAWSConfig  aws.Config
	ProjectName    string
	Route53Enabled bool
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var asgStatusResource controller.Resource
	{
		c := asgstatus.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		asgStatusResource, err = asgstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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

	var machineDeploymentResource controller.Resource
	{
		c := machinedeployment.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentResource, err = machinedeployment.New(c)
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

	resources := []controller.Resource{
		machineDeploymentResource,
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
		cr, err := key.ToCluster(obj)
		if err != nil {
			return false
		}

		if key.ClusterVersion(cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		cr, err := key.ToCluster(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var tenantClusterAWSClients aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient, cr)
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
