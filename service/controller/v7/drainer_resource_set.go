package v7

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"

	"github.com/giantswarm/aws-operator/client/aws"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v7/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v7/key"
	"github.com/giantswarm/aws-operator/service/controller/v7/resource/lifecycle"
)

type DrainerResourceSetConfig struct {
	GuestAWSClients aws.Clients
	Logger          micrologger.Logger

	GuestUpdateEnabled bool
	ProjectName        string
}

func NewDrainerResourceSet(config DrainerResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var cloudFormationService *cloudformationservice.CloudFormation
	{
		c := cloudformationservice.Config{
			Client: config.GuestAWSClients.CloudFormation,
		}

		cloudFormationService, err = cloudformationservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var lifecycleResource controller.CRUDResourceOps
	{
		c := lifecycle.ResourceConfig{
			Clients: config.GuestAWSClients,
			Logger:  config.Logger,
			Service: cloudFormationService,
		}

		lifecycleResource, err = lifecycle.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []controller.Resource
	ops := []controller.CRUDResourceOps{
		lifecycleResource,
	}
	for _, o := range ops {
		c := controller.CRUDResourceConfig{
			Logger: config.Logger,
			Ops:    o,
		}

		r, err := controller.NewCRUDResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		resources = append(resources, r)
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
