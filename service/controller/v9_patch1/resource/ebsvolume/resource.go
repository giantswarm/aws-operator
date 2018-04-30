package ebsvolume

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Name is the identifier of the resource.
	Name = "ebsvolumev9_patch1"
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
		logger:  config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func toEBSVolumeState(v interface{}) (*EBSVolumeState, error) {
	if v == nil {
		return nil, nil
	}

	volState, ok := v.(*EBSVolumeState)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", volState, v)
	}

	return volState, nil
}
