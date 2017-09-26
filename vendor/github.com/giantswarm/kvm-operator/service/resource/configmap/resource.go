package configmap

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/certificatetpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
)

const (
	KeyUserData = "user_data"
	// Name is the identifier of the resource.
	Name = "configmap"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	// Dependencies.
	CertWatcher certificatetpr.Searcher
	CloudConfig *cloudconfig.CloudConfig
	K8sClient   kubernetes.Interface
	Logger      micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new config map
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		CloudConfig: nil,
		K8sClient:   nil,
		Logger:      nil,
	}
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	certWatcher certificatetpr.Searcher
	cloudConfig *cloudconfig.CloudConfig
	k8sClient   kubernetes.Interface
	logger      micrologger.Logger
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		certWatcher: config.CertWatcher,
		cloudConfig: config.CloudConfig,
		k8sClient:   config.K8sClient,
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

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for config maps in the Kubernetes API")

	var configMaps []*apiv1.ConfigMap

	namespace := key.ClusterNamespace(customObject)
	configMapNames := key.ConfigMapNames(customObject)

	for _, name := range configMapNames {
		manifest, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a config map in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a config map in the Kubernetes API")
			configMaps = append(configMaps, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps in the Kubernetes API", len(configMaps)))

	return configMaps, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new config maps")

	configMaps, err := r.newConfigMaps(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new config maps", len(configMaps)))

	return configMaps, nil
}

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps have to be created")

	var configMapsToCreate []*apiv1.ConfigMap

	for _, desiredConfigMap := range desiredConfigMaps {
		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps that have to be created", len(configMapsToCreate)))

	return configMapsToCreate, nil
}

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps have to be deleted")

	var configMapsToDelete []*apiv1.ConfigMap

	for _, currentConfigMap := range currentConfigMaps {
		if containsConfigMap(desiredConfigMaps, currentConfigMap) {
			configMapsToDelete = append(configMapsToDelete, currentConfigMap)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps that have to be deleted", len(configMapsToDelete)))

	return configMapsToDelete, nil
}

func (r *Resource) GetUpdateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}

	var configMapsToCreate interface{}
	{
		configMapsToCreate, err = r.GetCreateState(ctx, obj, currentState, desiredState)
		if err != nil {
			return nil, nil, nil, microerror.Mask(err)
		}
	}

	var configMapsToDelete []*apiv1.ConfigMap
	{
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps have to be deleted")

		for _, currentConfigMap := range currentConfigMaps {
			if !containsConfigMap(desiredConfigMaps, currentConfigMap) {
				configMapsToDelete = append(configMapsToDelete, currentConfigMap)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps that have to be deleted", len(configMapsToDelete)))
	}

	var configMapsToUpdate []*apiv1.ConfigMap
	{
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps have to be updated")

		for _, currentConfigMap := range currentConfigMaps {
			desiredConfigMap, err := getConfigMapByName(desiredConfigMaps, currentConfigMap.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, nil, nil, microerror.Mask(err)
			}

			if isConfigMapModified(desiredConfigMap, currentConfigMap) {
				m, ok := messagecontext.FromContext(ctx)
				if ok {
					m.ConfigMapNames = append(m.ConfigMapNames, desiredConfigMap.Name)
				}
				configMapsToUpdate = append(configMapsToUpdate, desiredConfigMap)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps that have to be updated", len(configMapsToUpdate)))
	}

	return configMapsToCreate, configMapsToDelete, configMapsToUpdate, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToCreate, err := toConfigMaps(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create the config maps in the Kubernetes API.
	if len(configMapsToCreate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the config maps in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, configMap := range configMapsToCreate {
			_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(configMap)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the config maps in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the config maps do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToDelete, err := toConfigMaps(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMapsToDelete) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the config maps in the Kubernetes API")

		// Create the config maps in the Kubernetes API.
		namespace := key.ClusterNamespace(customObject)
		for _, configMap := range configMapsToDelete {
			err := r.k8sClient.CoreV1().ConfigMaps(namespace).Delete(configMap.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the config maps in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the config maps do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessUpdateState(ctx context.Context, obj, updateState interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToUpdate, err := toConfigMaps(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMapsToUpdate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "updating the config maps in the Kubernetes API")

		// Create the config maps in the Kubernetes API.
		namespace := key.ClusterNamespace(customObject)
		for _, configMap := range configMapsToUpdate {
			_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Update(configMap)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "updated the config maps in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the config maps do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

// newConfigMap creates a new Kubernetes configmap using the provided
// information. customObject is used for name and label creation. params serves
// as structure being injected into the template execution to interpolate
// variables. prefix can be either "master" or "worker" and is used to prefix
// the configmap name.
func (r *Resource) newConfigMap(customObject kvmtpr.CustomObject, template string, node clustertprspec.Node, prefix string) (*apiv1.ConfigMap, error) {
	var newConfigMap *apiv1.ConfigMap
	{
		newConfigMap = &apiv1.ConfigMap{
			ObjectMeta: apismetav1.ObjectMeta{
				Name: key.ConfigMapName(customObject, node, prefix),
				Labels: map[string]string{
					"cluster":  key.ClusterID(customObject),
					"customer": key.ClusterCustomer(customObject),
				},
			},
			Data: map[string]string{
				KeyUserData: template,
			},
		}
	}

	return newConfigMap, nil
}

func (r *Resource) newConfigMaps(customObject kvmtpr.CustomObject) ([]*apiv1.ConfigMap, error) {
	var configMaps []*apiv1.ConfigMap

	certs, err := r.certWatcher.SearchCerts(customObject.Spec.Cluster.Cluster.ID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, node := range customObject.Spec.Cluster.Masters {
		template, err := r.cloudConfig.NewMasterTemplate(customObject, certs, node)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customObject, template, node, key.MasterID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		template, err := r.cloudConfig.NewWorkerTemplate(customObject, certs, node)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customObject, template, node, key.WorkerID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func containsConfigMap(list []*apiv1.ConfigMap, item *apiv1.ConfigMap) bool {
	_, err := getConfigMapByName(list, item.Name)
	if err != nil {
		return false
	}

	return true
}

func getConfigMapByName(list []*apiv1.ConfigMap, name string) (*apiv1.ConfigMap, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isConfigMapModified(a, b *apiv1.ConfigMap) bool {
	return !reflect.DeepEqual(a.Data, b.Data)
}

func toConfigMaps(v interface{}) ([]*apiv1.ConfigMap, error) {
	if v == nil {
		return nil, nil
	}

	configMaps, ok := v.([]*apiv1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.ConfigMap{}, v)
	}

	return configMaps, nil
}
