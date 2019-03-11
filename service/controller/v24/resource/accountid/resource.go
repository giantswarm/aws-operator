package accountid

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/accountid"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
)

const (
	Name = "accountidv24"
)

type Config struct {
	Logger micrologger.Logger
	STS    STS
}

type Resource struct {
	logger micrologger.Logger

	accountID *accountid.AccountID
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.STS == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.STS must not be empty", config)
	}

	var err error

	var accountIDService *accountid.AccountID
	{
		c := accountid.Config{
			Logger: config.Logger,
			STS:    config.STS,
		}

		accountIDService, err = accountid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newResource := &Resource{
		logger: config.Logger,

		accountID: accountIDService,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addAccountIDToContext(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Here we take the STS client scoped to the control plane AWS account to
	// lookup its ID. The ID is then set to the controller context.
	{
		accountID, err := r.accountID.Lookup()
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.ControlPlane.AWSAccountID = accountID
	}

	// Here we take the STS client scoped to the tenant cluster AWS account to
	// lookup its ID. The ID is then set to the controller context.
	{
		var accountIDService *accountid.AccountID
		{
			c := accountid.Config{
				Logger: r.logger,
				STS:    cc.Client.TenantCluster.AWS.STS,
			}

			accountIDService, err = accountid.New(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		accountID, err := accountIDService.Lookup()
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.AWSAccountID = accountID
	}

	return nil
}
