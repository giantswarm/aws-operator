package accountid

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/internal/accountid"
)

const (
	Name = "accountid"
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

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
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
		r.logger.Debugf(ctx, "finding the control plane's AWS account ID")

		var accountIDService *accountid.AccountID
		{
			c := accountid.Config{
				Logger: r.logger,
				STS:    cc.Client.ControlPlane.AWS.STS,
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

		cc.Status.ControlPlane.AWSAccountID = accountID

		r.logger.Debugf(ctx, "found the control plane's AWS account ID %#q", accountID)
	}

	// Here we take the STS client scoped to the tenant cluster AWS account to
	// lookup its ID. The ID is then set to the controller context.
	{
		r.logger.Debugf(ctx, "finding the tenant cluster's AWS account ID")

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

		cc.Status.TenantCluster.AWS.AccountID = accountID

		r.logger.Debugf(ctx, "found the tenant cluster's AWS account ID %#q", accountID)
	}

	return nil
}
