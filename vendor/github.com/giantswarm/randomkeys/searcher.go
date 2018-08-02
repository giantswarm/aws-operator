package randomkeys

import (
	"fmt"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const (
	// watchTimeOut is the time to wait on watches against the Kubernetes API
	// before giving up and throwing an error.
	watchTimeOut = 90 * time.Second
)

type Config struct {
	// Dependencies.

	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

func NewSearcher(config Config) (*Searcher, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	s := &Searcher{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return s, nil
}

type Searcher struct {
	// Dependencies.

	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func (s *Searcher) SearchCluster(clusterID string) (Cluster, error) {
	var cluster Cluster

	keys := []struct {
		RandomKey *RandomKey
		Type      Key
	}{
		{RandomKey: &cluster.APIServerEncryptionKey, Type: EncryptionKey},
	}

	for _, k := range keys {
		err := s.search(k.RandomKey, clusterID, k.Type)
		if err != nil {
			return Cluster{}, microerror.Mask(err)
		}
	}

	return cluster, nil
}

func (s *Searcher) search(randomKey *RandomKey, clusterID string, key Key) error {
	// Select only secrets that match the given key and the given
	// cluster clusterID.
	selector := fmt.Sprintf("%s=%s, %s=%s", legacyRandomKeyLabel, key, legacyClusterIDLabel, clusterID)

	watcher, err := s.k8sClient.Core().Secrets(SecretNamespace).Watch(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return microerror.Maskf(executionError, "watching secrets, selector = %q: unexpected closed channel", selector)
			}

			switch event.Type {
			case watch.Added:
				err := fillRandomKeyFromSecret(randomKey, event.Object, clusterID, key)
				if err != nil {
					return microerror.Maskf(err, "watching secrets, selector = %q")
				}

				return nil
			case watch.Deleted:
				// Noop. Ignore deleted events. These are
				// handled by the certificate operator.
			case watch.Error:
				return microerror.Maskf(executionError, "watching secrets, selector = %q: %v", selector, apierrors.FromObject(event.Object))
			}
		case <-time.After(watchTimeOut):
			return microerror.Maskf(timeoutError, "waiting secrets, selector = %q", selector)
		}
	}
}

func fillRandomKeyFromSecret(randomkey *RandomKey, obj runtime.Object, clusterID string, key Key) error {
	secret, ok := obj.(*corev1.Secret)
	if !ok || secret == nil {
		return microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", secret, obj)
	}

	gotClusterID := secret.Labels[legacyClusterIDLabel]
	if clusterID != gotClusterID {
		return microerror.Maskf(invalidSecretError, "expected clusterID = %q, got %q", clusterID, gotClusterID)
	}
	gotKeys := secret.Labels[legacyRandomKeyLabel]
	if string(key) != gotKeys {
		return microerror.Maskf(invalidSecretError, "expected random key = %q, got %q", key, gotKeys)
	}
	if *randomkey, ok = secret.Data[string(EncryptionKey)]; !ok {
		return microerror.Maskf(invalidSecretError, "%q key missing", EncryptionKey)
	}

	return nil
}
