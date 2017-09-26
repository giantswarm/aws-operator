package healthz

import (
	"github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microendpoint/service/healthz/k8s"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new healthz
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	var err error

	var k8sService healthz.Service
	{
		k8sConfig := k8s.DefaultConfig()
		k8sConfig.K8sClient = config.K8sClient
		k8sConfig.Logger = config.Logger
		k8sService, err = k8s.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		K8s: k8sService,
	}

	return newService, nil
}

// Service is the healthz service collection.
type Service struct {
	K8s healthz.Service
}
