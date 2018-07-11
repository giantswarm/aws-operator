package service

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"reflect"
)

const (
	// Name is the identifier of the resource.
	Name = "servicev13"

	httpsPort         = 443
	masterServiceName = "master"
)

// Config represents the configuration used to create a new service resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource implements the service resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured service resource.
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
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func getServiceByName(list []*apiv1.Service, name string) (*apiv1.Service, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isServiceModified(a, b *apiv1.Service) bool {
	if !portsEqual(a, b) {
		return true
	}

	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	if !reflect.DeepEqual(a.Annotations, b.Annotations) {
		return true
	}

	return false
}

func toService(v interface{}) (*apiv1.Service, error) {
	if v == nil {
		return nil, nil
	}

	service, ok := v.(*apiv1.Service)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.Service{}, v)
	}

	return service, nil
}

// portsEqual is a function that is checking if ports in the service have same important values.
func portsEqual(a, b *apiv1.Service) bool {
	if a == nil || b == nil {
		return false
	}
	if len(a.Spec.Ports) != len(b.Spec.Ports) {
		return false
	}

	for i := 0; i < len(a.Spec.Ports); i++ {
		portA := a.Spec.Ports[i]
		portB := b.Spec.Ports[i]

		if portA.Name != portB.Name {
			return false
		}
		if !reflect.DeepEqual(portA.Port, portB.Port) {
			return false
		}
		if !reflect.DeepEqual(portA.Protocol, portB.Protocol) {
			return false
		}
	}
	return true
}
