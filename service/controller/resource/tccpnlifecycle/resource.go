package tccpnlifecycle

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpnlifecycle"
)

type Config struct {
	Logger micrologger.Logger
}

// Resource implements an operatorkit resource and provides a mechanism to fetch
// information from Cloud Formation stack outputs of the Tenant Cluster Control
// Plane Nodes stack.
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
