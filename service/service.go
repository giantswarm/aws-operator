// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"fmt"
	"sync"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"

	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
	"github.com/giantswarm/aws-operator/service/operator"
	"github.com/giantswarm/aws-operator/service/version"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Kubernetes.
	KubernetesAPIServer   string
	KubernetesUsername    string
	KubernetesPassword    string
	KubernetesBearerToken string
	KubernetesInsecure    bool

	Description string
	GitCommit   string
	Name        string
	Source      string
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Kubernetes.
		KubernetesAPIServer:   "",
		KubernetesUsername:    "",
		KubernetesPassword:    "",
		KubernetesBearerToken: "",
		KubernetesInsecure:    false,

		Description: "",
		GitCommit:   "",
		Name:        "",
		Source:      "",
	}
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating aws-operator service with config: %#v", config))

	var err error

	var k8sclient kubernetes.Interface
	{
		k8sclient, err = k8sutil.NewClient(
			config.KubernetesAPIServer,
			config.KubernetesUsername,
			config.KubernetesPassword,
			config.KubernetesBearerToken,
			config.KubernetesInsecure,
		)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var operatorService *operator.Service
	{
		operatorConfig := operator.DefaultConfig()

		operatorConfig.Logger = config.Logger
		operatorConfig.K8sclient = k8sclient

		operatorService, err = operator.New(operatorConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Description = config.Description
		versionConfig.GitCommit = config.GitCommit
		versionConfig.Name = config.Name
		versionConfig.Source = config.Source

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	newService := &Service{
		// Dependencies.
		Operator: operatorService,
		Version:  versionService,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	Operator *operator.Service
	Version  *version.Service

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Operator.Boot()
	})
}
