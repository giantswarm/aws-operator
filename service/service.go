// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/logresource"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/randomkeytpr"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/service/cloudconfig"
	"github.com/giantswarm/aws-operator/service/healthz"
	"github.com/giantswarm/aws-operator/service/messagecontext"
	"github.com/giantswarm/aws-operator/service/operator"
	cloudformationresource "github.com/giantswarm/aws-operator/service/resource/cloudformation"
	legacyresource "github.com/giantswarm/aws-operator/service/resource/legacy"
)

const (
	ResourceRetries uint64 = 3
	RedactedString         = "[REDACTED]"
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

type Service struct {
	// Dependencies.
	Healthz  *healthz.Service
	Operator *operator.Operator
	Version  *version.Service

	// Internals.
	bootOnce sync.Once
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()
		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.Logger = config.Logger
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultServiceConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keyWatcher *randomkeytpr.Service
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = k8sClient
		keyConfig.Logger = config.Logger
		keyWatcher, err = randomkeytpr.NewService(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsConfig awsclient.Config
	{
		awsConfig = awsclient.Config{
			AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
			AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
			SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
		}
	}

	var awsHostConfig awsclient.Config
	{
		accessKeyID := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID)
		accessKeySecret := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret)
		sessionToken := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session)

		if accessKeyID == "" && accessKeySecret == "" {
			config.Logger.Log("debug", "no host cluster account credentials supplied, assuming guest and host uses same account")
			awsHostConfig = awsConfig
		} else {
			config.Logger.Log("debug", "host cluster account credentials supplied, using separate accounts for host and guest clusters")
			awsHostConfig = awsclient.Config{
				AccessKeyID:     accessKeyID,
				AccessKeySecret: accessKeySecret,
				SessionToken:    sessionToken,
			}
		}
	}

	var ccService *cloudconfig.CloudConfig
	{
		ccServiceConfig := cloudconfig.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccService, err = cloudconfig.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyResource framework.Resource
	{
		legacyConfig := legacyresource.DefaultConfig()
		legacyConfig.AwsConfig = awsConfig
		legacyConfig.AwsHostConfig = awsHostConfig
		legacyConfig.CertWatcher = certWatcher
		legacyConfig.CloudConfig = ccService
		legacyConfig.K8sClient = k8sClient
		legacyConfig.KeyWatcher = keyWatcher
		legacyConfig.Logger = config.Logger
		legacyConfig.PubKeyFile = config.Viper.GetString(config.Flag.Service.AWS.PubKeyFile)

		legacyResource, err = legacyresource.New(legacyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudformationResource framework.Resource
	{
		cloudformationConfig := cloudformationresource.DefaultConfig()

		cloudformationConfig.CertWatcher = certWatcher
		//cloudformationConfig.CloudConfig = ccService
		cloudformationConfig.K8sClient = k8sClient
		cloudformationConfig.Logger = config.Logger

		cloudformationResource, err = cloudformationresource.New(cloudformationConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We create the list of resources and wrap each resource around some common
	// resources like metrics and retry resources.
	//
	// NOTE that the retry resources wrap the underlying resources first. The
	// wrapped resources are then wrapped around the metrics resource. That way
	// the metrics also consider execution times and execution attempts including
	// retries.
	var resources []framework.Resource
	{
		resources = []framework.Resource{
			legacyResource,
			cloudformationResource,
		}

		logWrapConfig := logresource.DefaultWrapConfig()
		logWrapConfig.Logger = config.Logger
		resources, err = logresource.Wrap(resources, logWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		retryWrapConfig := retryresource.DefaultWrapConfig()
		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger
		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()
		metricsWrapConfig.Name = config.Name
		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		ctx = messagecontext.NewContext(ctx, messagecontext.NewMessage())

		return ctx, nil
	}

	var frameworkBackOff *backoff.ExponentialBackOff
	{
		frameworkBackOff = backoff.NewExponentialBackOff()
		frameworkBackOff.MaxElapsedTime = 5 * time.Minute
	}

	var operatorFramework *framework.Framework
	{
		frameworkConfig := framework.DefaultConfig()

		frameworkConfig.BackOff = frameworkBackOff
		frameworkConfig.InitCtxFunc = initCtxFunc
		frameworkConfig.Logger = config.Logger
		frameworkConfig.Resources = resources

		operatorFramework, err = framework.New(frameworkConfig)
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

	var operatorBackOff *backoff.ExponentialBackOff
	{
		operatorBackOff = backoff.NewExponentialBackOff()
		operatorBackOff.MaxElapsedTime = 5 * time.Minute
	}

	var operatorService *operator.Operator
	{
		operatorConfig := operator.DefaultConfig()

		operatorConfig.BackOff = operatorBackOff
		operatorConfig.K8sClient = k8sClient
		operatorConfig.Logger = config.Logger
		operatorConfig.OperatorFramework = operatorFramework

		operatorService, err = operator.New(operatorConfig)
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
		Healthz:  healthzService,
		Operator: operatorService,
		Version:  versionService,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Operator.Boot()
	})
}
