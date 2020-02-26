package tccpnoutputs

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpnoutputs"
)

type Config struct {
	Logger micrologger.Logger

	Route53Enabled bool
}

// Resource implements an operatorkit resource and provides a mechanism to fetch
// information from Cloud Formation stack outputs of the Tenant Cluster Control
// Plane Node stack.
//
type Resource struct {
	logger micrologger.Logger

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
