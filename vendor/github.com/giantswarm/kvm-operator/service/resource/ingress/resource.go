package ingress

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

const (
	APIID  = "api"
	EtcdID = "etcd"
	// Name is the identifier of the resource.
	Name = "ingress"
)

// Config represents the configuration used to create a new ingress resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new ingress
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the ingress resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured ingress resource.
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
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for ingresses in the Kubernetes API")

	var ingresses []*v1beta1.Ingress

	namespace := key.ClusterNamespace(customObject)
	ingressNames := []string{
		APIID,
		EtcdID,
	}

	for _, name := range ingressNames {
		manifest, err := r.k8sClient.Extensions().Ingresses(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a ingress in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a ingress in the Kubernetes API")
			ingresses = append(ingresses, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d ingresses in the Kubernetes API", len(ingresses)))

	return ingresses, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new ingresses")

	var ingresses []*v1beta1.Ingress

	ingresses = append(ingresses, newAPIIngress(customObject))
	ingresses = append(ingresses, newEtcdIngress(customObject))

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new ingresses", len(ingresses)))

	return ingresses, nil
}

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentIngresses, err := toIngresses(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredIngresses, err := toIngresses(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which ingresses have to be created")

	var ingressesToCreate []*v1beta1.Ingress

	for _, desiredIngress := range desiredIngresses {
		if !containsIngress(currentIngresses, desiredIngress) {
			ingressesToCreate = append(ingressesToCreate, desiredIngress)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d ingresses that have to be created", len(ingressesToCreate)))

	return ingressesToCreate, nil
}

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentIngresses, err := toIngresses(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredIngresses, err := toIngresses(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which ingresses have to be deleted")

	var ingressesToDelete []*v1beta1.Ingress

	for _, currentIngress := range currentIngresses {
		if containsIngress(desiredIngresses, currentIngress) {
			ingressesToDelete = append(ingressesToDelete, currentIngress)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d ingresses that have to be deleted", len(ingressesToDelete)))

	return ingressesToDelete, nil
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
	ingressesToCreate, err := toIngresses(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(ingressesToCreate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the ingresses in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, ingress := range ingressesToCreate {
			_, err := r.k8sClient.Extensions().Ingresses(namespace).Create(ingress)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the ingresses in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the ingresses do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	ingressesToDelete, err := toIngresses(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(ingressesToDelete) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the ingresses in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, ingress := range ingressesToDelete {
			err := r.k8sClient.Extensions().Ingresses(namespace).Delete(ingress.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the ingresses in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the ingresses do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessUpdateState(ctx context.Context, obj, updateState interface{}) error {
	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsIngress(list []*v1beta1.Ingress, item *v1beta1.Ingress) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toIngresses(v interface{}) ([]*v1beta1.Ingress, error) {
	if v == nil {
		return nil, nil
	}

	ingresses, ok := v.([]*v1beta1.Ingress)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1beta1.Ingress{}, v)
	}

	return ingresses, nil
}
