package v8

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"

	"github.com/giantswarm/aws-operator/client/aws"
	cloudformationservice "github.com/giantswarm/aws-operator/service/awsconfig/v8/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v8/key"
	"github.com/giantswarm/aws-operator/service/awsconfig/v8/resource/lifecycle"
)

type DrainerResourceSetConfig struct {
	AWS    aws.Clients
	Logger micrologger.Logger

	GuestUpdateEnabled bool
	ProjectName        string
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*framework.ResourceSet, error) {
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

	var lifecycleResource framework.Resource
	{
		c := lifecycle.ResourceConfig{
			AWS:     config.AWS,
			Logger:  config.Logger,
			Service: cloudFormationService,
		}

		ops, err := lifecycle.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		lifecycleResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
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

	var resourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
