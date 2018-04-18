package v10

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"

	"github.com/giantswarm/aws-operator/client/aws"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v10/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v10/key"
	"github.com/giantswarm/aws-operator/service/controller/v10/resource/lifecycle"
)

type DrainerResourceSetConfig struct {
	AWS       aws.Clients
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	GuestUpdateEnabled bool
	ProjectName        string
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var cloudFormationService *cloudformationservice.CloudFormation
	{
		c := cloudformationservice.Config{
			Client: config.AWS.CloudFormation,
		}

		cloudFormationService, err = cloudformationservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var lifecycleResource controller.Resource
	{
		c := lifecycle.ResourceConfig{
			AWS:       config.AWS,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
			Service:   cloudFormationService,
		}

		lifecycleResource, err = lifecycle.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		lifecycleResource,
	}

	{
		c := retryresource.WrapConfig{
			BackOffFactory: func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), uint64(3)) },
			Logger:         config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

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
