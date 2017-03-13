// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"fmt"
	"sync"

	"github.com/ericchiang/k8s"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
	"github.com/giantswarm/aws-operator/service/create"
	"github.com/giantswarm/aws-operator/service/version"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Sub-dependencies configs.
	AwsConfig awsutil.Config
	K8sConfig k8sutil.Config

	// AWS cerfificates options.
	CertsDir string

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

		// Sub-dependencies configs.
		AwsConfig: awsutil.Config{},
		K8sConfig: k8sutil.Config{},

		// AWS certificates optionts.
		CertsDir: "",

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

	var k8sClient *k8s.Client
	{
		k8sClient, err = k8sutil.NewClient(config.K8sConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var createService *create.Service
	{
		createConfig := create.DefaultConfig()

		createConfig.AwsConfig = config.AwsConfig
		createConfig.K8sClient = k8sClient
		createConfig.Logger = config.Logger
		createConfig.CertsDir = config.CertsDir

		createService, err = create.New(createConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}
	fmt.Println(config.AwsConfig)

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
		Create:  createService,
		Version: versionService,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	Create  *create.Service
	Version *version.Service

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Create.Boot()
	})
}
