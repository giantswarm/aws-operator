package endpoints

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "endpointsv6"

	httpsPort           = 443
	masterEndpointsName = "master"
)

// Config represents the configuration used to create a new endpoints resource.
type Config struct {
	// Dependencies.
	Clients   Clients
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource implements the endpoints resource.
type Resource struct {
	// Dependencies.
	awsClients Clients
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

// New creates a new configured endpoints resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if reflect.DeepEqual(config.Clients, Clients{}) {
		return nil, microerror.Maskf(invalidConfigError, "config.Clients must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		awsClients: config.Clients,
		k8sClient:  config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func toEndpoints(v interface{}) (*apiv1.Endpoints, error) {
	if v == nil {
		return nil, nil
	}

	endpoints, ok := v.(*apiv1.Endpoints)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.Endpoints{}, v)
	}

	return endpoints, nil
}
