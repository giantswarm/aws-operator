package loadbalancer

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "loadbalancerv5"
)

// Config represents the configuration used to create a new loadbalancer resource.
type Config struct {
	// Dependencies.
	Clients Clients
	Logger  micrologger.Logger
}

// Resource implements the loadbalancer resource.
type Resource struct {
	// Dependencies.
	clients Clients
	logger  micrologger.Logger
}

// New creates a new configured loadbalancer resource.
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

func toLoadBalancerState(v interface{}) (*LoadBalancerState, error) {
	if v == nil {
		return nil, nil
	}

	lbState, ok := v.(*LoadBalancerState)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", lbState, v)
	}

	return lbState, nil
}
