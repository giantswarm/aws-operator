package loadbalancer

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Name is the identifier of the resource.
	Name = "loadbalancerv17"
)

// Config represents the configuration used to create a new loadbalancer resource.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// Resource implements the loadbalancer resource.
type Resource struct {
	// Dependencies.
	logger micrologger.Logger
}

// New creates a new configured loadbalancer resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		logger: config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
