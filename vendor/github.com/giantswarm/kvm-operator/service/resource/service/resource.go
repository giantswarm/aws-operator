package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

const (
	// Name is the identifier of the resource.
	Name = "service"
)

// Config represents the configuration used to create a new service resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new service
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
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

	newService := &Resource{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for services in the Kubernetes API")

	var services []*apiv1.Service

	namespace := key.ClusterNamespace(customObject)
	serviceNames := []string{
		key.MasterID,
		key.WorkerID,
	}

	for _, name := range serviceNames {
		manifest, err := r.k8sClient.CoreV1().Services(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a service in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a service in the Kubernetes API")
			services = append(services, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d services in the Kubernetes API", len(services)))

	return services, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new services")

	var services []*apiv1.Service

	services = append(services, newMasterService(customObject))
	services = append(services, newWorkerService(customObject))

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new services", len(services)))

	return services, nil
}

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentServices, err := toServices(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServices, err := toServices(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which services have to be created")

	var servicesToCreate []*apiv1.Service

	for _, desiredService := range desiredServices {
		if !containsService(currentServices, desiredService) {
			servicesToCreate = append(servicesToCreate, desiredService)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d services that have to be created", len(servicesToCreate)))

	return servicesToCreate, nil
}

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentServices, err := toServices(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServices, err := toServices(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which services have to be deleted")

	var servicesToDelete []*apiv1.Service

	for _, currentService := range currentServices {
		if containsService(desiredServices, currentService) {
			servicesToDelete = append(servicesToDelete, currentService)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d services that have to be deleted", len(servicesToDelete)))

	return servicesToDelete, nil
}

func (r *Resource) GetUpdateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	return nil, nil, nil, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	servicesToCreate, err := toServices(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(servicesToCreate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the services in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, service := range servicesToCreate {
			_, err := r.k8sClient.CoreV1().Services(namespace).Create(service)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the services in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the services do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	servicesToDelete, err := toServices(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(servicesToDelete) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the services in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, service := range servicesToDelete {
			err := r.k8sClient.CoreV1().Services(namespace).Delete(service.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the services in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the services do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessUpdateState(ctx context.Context, obj, updateState interface{}) error {
	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsService(list []*apiv1.Service, item *apiv1.Service) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toServices(v interface{}) ([]*apiv1.Service, error) {
	if v == nil {
		return nil, nil
	}

	services, ok := v.([]*apiv1.Service)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.Service{}, v)
	}

	return services, nil
}
