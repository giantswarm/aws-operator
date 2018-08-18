package setup

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	AWSClient *awsclient.Client
	Guest     *framework.Guest
	Host      *framework.Host
	Logger    micrologger.Logger
}

func (c Config) Validate() error {
	if c.AWSClient == nil {
		return microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", c)
	}
	if c.Guest == nil {
		return microerror.Maskf(invalidConfigError, "%T.Guest must not be empty", c)
	}
	if c.Host == nil {
		return microerror.Maskf(invalidConfigError, "%T.Host must not be empty", c)
	}
	if c.Logger == nil {
		return microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	return nil
}
