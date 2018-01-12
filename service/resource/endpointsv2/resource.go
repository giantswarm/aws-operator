package endpointsv2

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	// Name is the identifier of the resource.
	Name = "endpointsv2"

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

// DefaultConfig provides a default configuration to create a new endpoints
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Clients:   Clients{},
		K8sClient: nil,
		Logger:    nil,
	}
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

func (r *Resource) Underlying() framework.Resource {
	return r
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
