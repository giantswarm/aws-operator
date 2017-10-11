package randomkeytpr

import (
	"fmt"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

const (
	// WatchTimeOut is the time to wait on watches against the Kubernetes API
	// before giving up and throwing an error.
	WatchTimeOut = 90 * time.Second
)

// ServiceConfig represents the configuration used to create a certificate TPR
// service.
type ServiceConfig struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultServiceConfig provides a default configuration to create a new
// certificate TPR service by best effort.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// NewService creates a new configured certificate TPR service.
func NewService(config ServiceConfig) (*Service, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return newService, nil
}

// Service implements the certificate TPR service.
type Service struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// SearchKeys watches for keys secrets of a cluster
func (s *Service) SearchKeys(clusterID string) (map[Key][]byte, error) {
	keys := make(map[Key][]byte)

	for _, keyType := range RandomKeyTypes {
		ab, err := s.SearchKeysForKeytype(clusterID, keyType.String())
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for k, v := range ab {
			keys[k] = v
		}
	}

	return keys, nil
}

// SearchKeysForKeytype watches for keys secrets of a single cluster keytype and
// returns it as assets bundle.
func (s *Service) SearchKeysForKeytype(clusterID, keyType string) (map[Key][]byte, error) {
	// TODO we should also do a list. In case the secrets have already been
	// created we might miss them with only watching.
	s.logger.Log("debug", fmt.Sprintf("searching secret: %s=%s, %s=%s", KeyLabel, keyType, ClusterIDLabel, clusterID))

	watcher, err := s.k8sClient.Core().Secrets(api.NamespaceDefault).Watch(apismetav1.ListOptions{
		// Select only secrets that match the given Keytype and the given cluster
		// clusterID.
		LabelSelector: fmt.Sprintf(
			"%s=%s, %s=%s",
			KeyLabel,
			keyType,
			ClusterIDLabel,
			clusterID,
		),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	keys := make(map[Key][]byte)

	defer watcher.Stop()
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil, microerror.Maskf(secretsRetrievalFailedError, "secrets channel was already closed")
			}

			switch event.Type {
			case watch.Added:
				secret := event.Object.(*v1.Secret)
				key := Key(secret.Labels[KeyLabel])

				if !ValidKey(key) {
					return nil, microerror.Maskf(secretsRetrievalFailedError, "unknown clusterKey %s", key)
				}

				for _, k := range RandomKeyTypes {
					asset, ok := secret.Data[k.String()]
					if !ok {
						return nil, microerror.Maskf(secretsRetrievalFailedError, "malformed secret was missing %v asset", keyType)
					}

					keys[k] = asset
				}

				return keys, nil
			case watch.Deleted:
				// Noop. Ignore deleted events. These are handled by the certificate
				// operator.
			case watch.Error:
				return nil, microerror.Maskf(secretsRetrievalFailedError, "there was an error in the watcher: %v", apierrors.FromObject(event.Object))
			}
		case <-time.After(WatchTimeOut):
			return nil, microerror.Maskf(secretsRetrievalFailedError, "timed out waiting for secrets")
		}
	}
}
