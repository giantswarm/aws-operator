package deployment

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/updateallowedcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
)

const (
	// Name is the identifier of the resource.
	Name = "deployment"
)

// Config represents the configuration used to create a new deployment resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new deployment
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the deployment resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured deployment resource.
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

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for deployments in the Kubernetes API")

	var deployments []*v1beta1.Deployment

	namespace := key.ClusterNamespace(customObject)
	deploymentNames := key.DeploymentNames(customObject)

	for _, name := range deploymentNames {
		manifest, err := r.k8sClient.Extensions().Deployments(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a deployment in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a deployment in the Kubernetes API")
			deployments = append(deployments, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments in the Kubernetes API", len(deployments)))

	return deployments, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new deployments")

	var deployments []*v1beta1.Deployment

	{
		masterDeployments, err := newMasterDeployments(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, masterDeployments...)

		workerDeployments, err := newWorkerDeployments(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, workerDeployments...)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new deployments", len(deployments)))

	return deployments, nil
}

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which deployments have to be created")

	var deploymentsToCreate []*v1beta1.Deployment

	for _, desiredDeployment := range desiredDeployments {
		if !containsDeployment(currentDeployments, desiredDeployment) {
			deploymentsToCreate = append(deploymentsToCreate, desiredDeployment)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments that have to be created", len(deploymentsToCreate)))

	return deploymentsToCreate, nil
}

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which deployments have to be deleted")

	var deploymentsToDelete []*v1beta1.Deployment

	for _, currentDeployment := range currentDeployments {
		if containsDeployment(desiredDeployments, currentDeployment) {
			deploymentsToDelete = append(deploymentsToDelete, currentDeployment)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments that have to be deleted", len(deploymentsToDelete)))

	return deploymentsToDelete, nil
}

func (r *Resource) GetUpdateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}

	var deploymentsToCreate interface{}
	{
		deploymentsToCreate, err = r.GetCreateState(ctx, obj, currentState, desiredState)
		if err != nil {
			return nil, nil, nil, microerror.Mask(err)
		}
	}

	var deploymentsToDelete []*v1beta1.Deployment
	{
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which deployments have to be deleted")

		for _, currentDeployment := range currentDeployments {
			if !containsDeployment(desiredDeployments, currentDeployment) {
				deploymentsToDelete = append(deploymentsToDelete, currentDeployment)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments that have to be deleted", len(deploymentsToDelete)))
	}

	var deploymentsToUpdate []*v1beta1.Deployment
	if updateallowedcontext.IsUpdateAllowed(ctx) {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which deployments have to be updated")

		// Check if config maps of deployments changed. In case they did, add the
		// deployments to the list of deployments intended to be updated.
		m, ok := messagecontext.FromContext(ctx)
		if ok {
			for _, name := range m.ConfigMapNames {
				desiredDeployment, err := getDeploymentByConfigMapName(desiredDeployments, name)
				if err != nil {
					return nil, nil, nil, microerror.Mask(err)
				}
				deploymentsToUpdate = append(deploymentsToUpdate, desiredDeployment)
			}
		}

		// Check if deployments changed. In case they did, add the deployments to
		// the list of deployments intended to be updated, but only in case they are
		// not already being tracked.
		for _, currentDeployment := range currentDeployments {
			desiredDeployment, err := getDeploymentByName(desiredDeployments, currentDeployment.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, nil, nil, microerror.Mask(err)
			}

			if !isDeploymentModified(desiredDeployment, currentDeployment) {
				continue
			}

			if containsDeployment(deploymentsToUpdate, desiredDeployment) {
				continue
			}

			deploymentsToUpdate = append(deploymentsToUpdate, desiredDeployment)
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments that have to be updated", len(deploymentsToUpdate)))
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "not computing update state because deployments are not allowed to be updated")
	}

	return deploymentsToCreate, deploymentsToDelete, deploymentsToUpdate, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToCreate, err := toDeployments(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToCreate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToCreate {
			_, err := r.k8sClient.Extensions().Deployments(namespace).Create(deployment)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the deployments in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the deployments do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToDelete, err := toDeployments(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToDelete) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToDelete {
			err := r.k8sClient.Extensions().Deployments(namespace).Delete(deployment.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the deployments in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the deployments do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessUpdateState(ctx context.Context, obj, updateState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToUpdate, err := toDeployments(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToUpdate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "updating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToUpdate {
			_, err := r.k8sClient.Extensions().Deployments(namespace).Update(deployment)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "updated the deployments in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the deployments do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsDeployment(list []*v1beta1.Deployment, item *v1beta1.Deployment) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getDeploymentByName(list []*v1beta1.Deployment, name string) (*v1beta1.Deployment, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func getDeploymentByConfigMapName(list []*v1beta1.Deployment, name string) (*v1beta1.Deployment, error) {
	for _, l := range list {
		for _, v := range l.Spec.Template.Spec.Volumes {
			if v.VolumeSource.ConfigMap.LocalObjectReference.Name == name {
				return l, nil
			}
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isDeploymentModified(a, b *v1beta1.Deployment) bool {
	return !reflect.DeepEqual(a.Spec.Template.Spec, b.Spec.Template.Spec)
}

func toDeployments(v interface{}) ([]*v1beta1.Deployment, error) {
	if v == nil {
		return nil, nil
	}

	deployments, ok := v.([]*v1beta1.Deployment)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1beta1.Deployment{}, v)
	}

	return deployments, nil
}
