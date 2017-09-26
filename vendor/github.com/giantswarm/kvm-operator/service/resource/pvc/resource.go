package pvc

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
	Name = "pvc"
	// StorageClass is the storage class annotation persistent volume claims are
	// configured with.
	StorageClass = "g8s-storage"
)

// Config represents the configuration used to create a new PVC resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new PVC
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the PVC resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured PVC resource.
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

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for PVCs in the Kubernetes API")

	var PVCs []*apiv1.PersistentVolumeClaim

	namespace := key.ClusterNamespace(customObject)
	pvcNames := key.PVCNames(customObject)

	for _, name := range pvcNames {
		manifest, err := r.k8sClient.Core().PersistentVolumeClaims(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a PVC in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a PVC in the Kubernetes API")
			PVCs = append(PVCs, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d PVCs in the Kubernetes API", len(PVCs)))

	return PVCs, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var PVCs []*apiv1.PersistentVolumeClaim

	if key.StorageType(customObject) == "persistentVolume" {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new PVCs")

		PVCs, err = newEtcdPVCs(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new PVCs", len(PVCs)))
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "not computing the new PVCs because storage type is not 'persistentVolume'")
	}

	return PVCs, nil
}

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentPVCs, err := toPVCs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredPVCs, err := toPVCs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which PVCs have to be created")

	var pvcsToCreate []*apiv1.PersistentVolumeClaim

	for _, desiredPVC := range desiredPVCs {
		if !containsPVC(currentPVCs, desiredPVC) {
			pvcsToCreate = append(pvcsToCreate, desiredPVC)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d PVCs that have to be created", len(pvcsToCreate)))

	return pvcsToCreate, nil
}

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentPVCs, err := toPVCs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredPVCs, err := toPVCs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which PVCs have to be deleted")

	var pvcsToDelete []*apiv1.PersistentVolumeClaim

	for _, currentPVC := range currentPVCs {
		if containsPVC(desiredPVCs, currentPVC) {
			pvcsToDelete = append(pvcsToDelete, currentPVC)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d PVCs that have to be deleted", len(pvcsToDelete)))

	return pvcsToDelete, nil
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
	pvcsToCreate, err := toPVCs(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(pvcsToCreate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the PVCs in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, PVC := range pvcsToCreate {
			_, err := r.k8sClient.Core().PersistentVolumeClaims(namespace).Create(PVC)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the PVCs in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the PVCs do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	pvcsToDelete, err := toPVCs(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(pvcsToDelete) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the PVCs in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, PVC := range pvcsToDelete {
			err := r.k8sClient.Core().PersistentVolumeClaims(namespace).Delete(PVC.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the PVCs in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the PVCs do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessUpdateState(ctx context.Context, obj, updateState interface{}) error {
	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsPVC(list []*apiv1.PersistentVolumeClaim, item *apiv1.PersistentVolumeClaim) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toPVCs(v interface{}) ([]*apiv1.PersistentVolumeClaim, error) {
	if v == nil {
		return nil, nil
	}

	PVCs, ok := v.([]*apiv1.PersistentVolumeClaim)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.PersistentVolumeClaim{}, v)
	}

	return PVCs, nil
}
