package stackoutput

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "stackoutputv24"
)

type Config struct {
	EC2    EC2
	Logger micrologger.Logger

	Route53Enabled bool
}

type Resource struct {
	ec2    EC2
	logger micrologger.Logger

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if config.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.EC2 must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		ec2:    config.EC2,
		logger: config.Logger,

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
