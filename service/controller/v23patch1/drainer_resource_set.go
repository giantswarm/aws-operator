package v23patch1

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v23patch1/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/credential"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/key"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/resource/drainer"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/resource/drainfinisher"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/resource/stackoutput"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/resource/workerasgname"
)

type DrainerResourceSetConfig struct {
	G8sClient     versioned.Interface
	HostAWSConfig aws.Config
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger

	GuestUpdateEnabled bool
	ProjectName        string
	Route53Enabled     bool
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

	var stackOutputResource controller.Resource
	{
		c := stackoutput.Config{
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		stackOutputResource, err = stackoutput.New(c)
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
		stackOutputResource,
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
		if config.GuestUpdateEnabled {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}

		var awsClient aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient, obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			c := config.HostAWSConfig
			c.RoleARN = arn

			awsClient, err = aws.NewClients(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		var cloudFormationService *cloudformationservice.CloudFormation
		{
			c := cloudformationservice.Config{
				Client: awsClient.CloudFormation,
			}

			cloudFormationService, err = cloudformationservice.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		c := controllercontext.Context{
			AWSClient:      awsClient,
			CloudFormation: *cloudFormationService,
		}
		ctx = controllercontext.NewContext(ctx, c)

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
