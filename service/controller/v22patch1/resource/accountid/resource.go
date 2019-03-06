package accountid

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v22patch1/controllercontext"
)

const (
	Name = "accountidv22patch1"
)

type Config struct {
	Logger micrologger.Logger
}

type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newResource := &Resource{
		logger: config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func addAccountIDToContext(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	accountID, err := cc.AWSService.GetAccountID()
	if err != nil {
		return microerror.Mask(err)
	}

	cc.Status.Cluster.AWSAccount.ID = accountID

	return nil
}
