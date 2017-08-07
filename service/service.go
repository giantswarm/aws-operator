// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"fmt"
	"sync"

	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	k8sclient "github.com/giantswarm/aws-operator/client/k8s"
	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/service/create"
	"github.com/giantswarm/aws-operator/service/healthz"
	"github.com/giantswarm/aws-operator/service/version"
)

const (
	RedactedString = "[REDACTED]"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

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

		// Settings.
		Flag:  nil,
		Viper: nil,

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
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating aws-operator service with config: %s", config))

	var err error

	// TODO this should come from operatorkit
	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.Config{
			InCluster:   config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			Host:        config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			Username:    config.Viper.GetString(config.Flag.Service.Kubernetes.Username),
			Password:    config.Viper.GetString(config.Flag.Service.Kubernetes.Password),
			BearerToken: config.Viper.GetString(config.Flag.Service.Kubernetes.BearerToken),
			TLSClientConfig: k8sclient.TLSClientConfig{
				CAFile:   config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CertFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		k8sClient, err = k8sutil.NewClient(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.New(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsConfig awsclient.Config
	{
		awsConfig = awsclient.Config{
			AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
			AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
		}
	}

	var createService *create.Service
	{
		createConfig := create.DefaultConfig()
		createConfig.AwsConfig = awsConfig
		createConfig.CertWatcher = certWatcher
		createConfig.K8sClient = k8sClient
		createConfig.Logger = config.Logger
		createConfig.PubKeyFile = config.Viper.GetString(config.Flag.Service.AWS.PubKeyFile)

		createService, err = create.New(createConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var healthzService *healthz.Service
	{
		healthzConfig := healthz.DefaultConfig()
		healthzConfig.AwsConfig = awsConfig
		healthzConfig.Logger = config.Logger

		healthzService, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Dependencies.
		Create:  createService,
		Healthz: healthzService,
		Version: versionService,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	Create  *create.Service
	Healthz *healthz.Service
	Version *version.Service

	// Internals.
	bootOnce sync.Once
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Create.Boot()
	})
}
