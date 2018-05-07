package ebsvolume

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v9patch1/ebs"
)

const (
	// Name is the identifier of the resource.
	Name = "ebsvolumev9patch1"
)

// Config represents the configuration used to create a new ebsvolume resource.
type Config struct {
	Logger  micrologger.Logger
	Service ebs.Interface
}

// Resource implements the ebsvolume resource.
type Resource struct {
	logger  micrologger.Logger
	service ebs.Interface
}

// New creates a new configured ebsvolume resource.
func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Service must not be empty", config)
	}

	newResource := &Resource{
		// Dependencies.
		logger:  config.Logger,
		service: config.Service,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
