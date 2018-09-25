package ebsvolume

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Name is the identifier of the resource.
	Name = "ebsvolumev16patch1"
)

// Config represents the configuration used to create a new ebsvolume resource.
type Config struct {
	Logger micrologger.Logger
}

// Resource implements the ebsvolume resource.
type Resource struct {
	logger micrologger.Logger
}

// New creates a new configured ebsvolume resource.
func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
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
