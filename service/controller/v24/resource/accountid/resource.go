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

func (r *Resource) addAccountIDToContext(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var s *accountid.AccountID
	{
		c := accountid.Config{
			Logger: r.logger,
			STS:    cc.AWSClient.STS,
		}

		s, err = accountid.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	accountID, err := s.Lookup()
	if err != nil {
		return microerror.Mask(err)
	}

	cc.Status.TenantCluster.AWSAccountID = accountID

	return nil
}
