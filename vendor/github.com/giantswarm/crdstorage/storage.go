package crdstorage

import (
	"context"
	"strings"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	CRDClient *k8scrdclient.CRDClient
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	CRD       *apiextensionsv1beta1.CustomResourceDefinition
	Name      string
	Namespace *corev1.Namespace
}

func DefaultConfig() Config {
	return Config{
		CRDClient: nil,
		G8sClient: nil,
		K8sClient: nil,
		Logger:    nil,

		CRD:       v1alpha1.NewStorageConfigCRD(),
		Name:      "",
		Namespace: nil,
	}
}

type Storage struct {
	crdClient *k8scrdclient.CRDClient
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	crd       *apiextensionsv1beta1.CustomResourceDefinition
	name      string
	namespace *corev1.Namespace
}

// New creates an uninitialized instance of Storage. It is required to call Boot
// before running any read/write operations against the returned Storage
// instance.
func New(config Config) (*Storage, error) {
	if config.CRDClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CRDClient must not be empty")
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	if config.CRD == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CRD must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}
	if config.Namespace == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Namespace must not be empty")
	}

	s := &Storage{
		crdClient: config.CRDClient,
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger: config.Logger.With(
			"crdName", config.CRD.GetName(),
			"crdVersion", config.CRD.GroupVersionKind(),
		),

		crd:       config.CRD,
		name:      config.Name,
		namespace: config.Namespace,
	}

	return s, nil
}

// Boot initializes the Storage by ensuring Kubernetes resources used by the
// Storage are in place. It is safe to call Boot more than once.
func (s *Storage) Boot(ctx context.Context) error {
	// Create CRD.
	{
		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 0
		backOff := backoff.WithMaxTries(b, 7)

		err := s.crdClient.EnsureCreated(ctx, s.crd, backOff)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Create namespace.
	{
		_, err := s.k8sClient.CoreV1().Namespaces().Create(s.namespace)
		if errors.IsAlreadyExists(err) {
			// TODO logs
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			// TODO logs
		}
	}

	// Create CRO.
	{
		storageConfig := &v1alpha1.StorageConfig{}

		storageConfig.Kind = "StorageConfig"
		storageConfig.APIVersion = "core.giantswarm.io/v1alpha1"
		storageConfig.Name = s.name
		storageConfig.Namespace = s.namespace.Name
		storageConfig.Spec.Storage.Data = map[string]string{}

		operation := func() error {
			_, err := s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Create(storageConfig)
			if errors.IsAlreadyExists(err) {
				// TODO logs
			} else if err != nil {
				return microerror.Mask(err)
			} else {
				// TODO logs
			}

			return nil
		}

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 0
		backOff := backoff.WithMaxTries(b, 7)

		err := backoff.Retry(operation, backOff)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (s *Storage) Delete(ctx context.Context, k microstorage.K) error {
	storageConfig, err := s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Get(s.name, apismetav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	delete(storageConfig.Spec.Storage.Data, k.Key())

	_, err = s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Update(storageConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (s *Storage) Exists(ctx context.Context, k microstorage.K) (bool, error) {
	storageConfig, err := s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Get(s.name, apismetav1.GetOptions{})
	if err != nil {
		return false, microerror.Mask(err)
	}

	_, ok := storageConfig.Spec.Storage.Data[k.Key()]
	if ok {
		return true, nil
	}

	return false, nil
}

func (s *Storage) List(ctx context.Context, k microstorage.K) ([]microstorage.KV, error) {
	storageConfig, err := s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Get(s.name, apismetav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	key := k.Key()

	// Special case.
	if key == "/" {
		var list []microstorage.KV
		for k, v := range storageConfig.Spec.Storage.Data {
			list = append(list, microstorage.MustKV(microstorage.NewKV(k, v)))
		}
		return list, nil
	}

	var list []microstorage.KV

	keyLen := len(key)
	for k, v := range storageConfig.Spec.Storage.Data {
		if len(k) <= keyLen+1 {
			continue
		}
		if !strings.HasPrefix(k, key) {
			continue
		}

		// k must be exact match or be separated with /.
		// I.e. /foo is under /foo/bar but not under /foobar.
		if k[keyLen] != '/' {
			continue
		}

		k = k[keyLen+1:]
		list = append(list, microstorage.MustKV(microstorage.NewKV(k, v)))
	}

	return list, nil
}

func (s *Storage) Put(ctx context.Context, kv microstorage.KV) error {
	storageConfig, err := s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Get(s.name, apismetav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	storageConfig.Spec.Storage.Data[kv.Key()] = kv.Val()

	_, err = s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Update(storageConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (s *Storage) Search(ctx context.Context, k microstorage.K) (microstorage.KV, error) {
	storageConfig, err := s.g8sClient.CoreV1alpha1().StorageConfigs(s.namespace.Name).Get(s.name, apismetav1.GetOptions{})
	if err != nil {
		return microstorage.KV{}, microerror.Mask(err)
	}

	key := k.Key()
	value, ok := storageConfig.Spec.Storage.Data[key]
	if ok {
		return microstorage.MustKV(microstorage.NewKV(key, value)), nil
	}

	return microstorage.KV{}, microerror.Maskf(notFoundError, "no value for key '%s'", key)
}
