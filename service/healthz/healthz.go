package healthz

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microendpoint/service/healthz/k8s"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	healthziam "github.com/giantswarm/aws-operator/service/healthz/iam"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	// Dependencies.
	IAMClient *iam.IAM
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new healthz
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		IAMClient: nil,
		K8sClient: nil,
		Logger:    nil,
	}
}

// Service is the healthz service collection.
type Service struct {
	IAM healthz.Service
	K8s healthz.Service
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	var err error

	var iamService healthz.Service
	{
		iamConfig := healthziam.DefaultConfig()
		iamConfig.IAMClient = config.IAMClient
		iamConfig.Logger = config.Logger
		iamService, err = healthziam.New(iamConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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
		IAM: iamService,
		K8s: k8sService,
	}

	return newService, nil
}
