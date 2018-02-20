package ebsvolume

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "ebsvolumev5"
)

// Config represents the configuration used to create a new ebsvolume resource.
type Config struct {
	// Dependencies.
	Clients Clients
	Logger  micrologger.Logger
}

// Resource implements the ebsvolume resource.
type Resource struct {
	// Dependencies.
	clients Clients
	logger  micrologger.Logger
}

// New creates a new configured ebsvolume resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if reflect.DeepEqual(config.Clients, Clients{}) {
		return nil, microerror.Maskf(invalidConfigError, "config.Clients must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		clients: config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
