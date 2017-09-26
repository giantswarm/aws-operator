// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8s"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/logresource"
	"github.com/giantswarm/operatorkit/framework/metricsresource"
	"github.com/giantswarm/operatorkit/framework/retryresource"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/kvm-operator/service/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/healthz"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
	"github.com/giantswarm/kvm-operator/service/operator"
	configmapresource "github.com/giantswarm/kvm-operator/service/resource/configmap"
	deploymentresource "github.com/giantswarm/kvm-operator/service/resource/deployment"
	ingressresource "github.com/giantswarm/kvm-operator/service/resource/ingress"
	namespaceresource "github.com/giantswarm/kvm-operator/service/resource/namespace"
	pvcresource "github.com/giantswarm/kvm-operator/service/resource/pvc"
	serviceresource "github.com/giantswarm/kvm-operator/service/resource/service"
)

const (
	ResourceRetries uint64 = 3
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
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating kvm-operator with config: %#v", config))

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8s.DefaultConfig()
		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.Logger = config.Logger
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8s.NewClient(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher certificatetpr.Searcher
	{
		certConfig := certificatetpr.DefaultServiceConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
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

	var configMapResource framework.Resource
	{
		configMapConfig := configmapresource.DefaultConfig()

		configMapConfig.CertWatcher = certWatcher
		configMapConfig.CloudConfig = ccService
		configMapConfig.K8sClient = k8sClient
		configMapConfig.Logger = config.Logger

		configMapResource, err = configmapresource.New(configMapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		deploymentConfig := deploymentresource.DefaultConfig()

		deploymentConfig.K8sClient = k8sClient
		deploymentConfig.Logger = config.Logger

		deploymentResource, err = deploymentresource.New(deploymentConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource framework.Resource
	{
		ingressConfig := ingressresource.DefaultConfig()

		ingressConfig.K8sClient = k8sClient
		ingressConfig.Logger = config.Logger

		ingressResource, err = ingressresource.New(ingressConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		namespaceConfig := namespaceresource.DefaultConfig()

		namespaceConfig.K8sClient = k8sClient
		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespaceresource.New(namespaceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource framework.Resource
	{
		pvcConfig := pvcresource.DefaultConfig()

		pvcConfig.K8sClient = k8sClient
		pvcConfig.Logger = config.Logger

		pvcResource, err = pvcresource.New(pvcConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		serviceConfig := serviceresource.DefaultConfig()

		serviceConfig.K8sClient = k8sClient
		serviceConfig.Logger = config.Logger

		serviceResource, err = serviceresource.New(serviceConfig)
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
	//
	// NOTE that the order of the namespace resource is important. We have to
	// start the namespace resource at first because all other resources are
	// created within this namespace.
	var resources []framework.Resource
	{
		resources = []framework.Resource{
			namespaceResource,

			configMapResource,
			deploymentResource,
			ingressResource,
			pvcResource,
			serviceResource,
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
		metricsWrapConfig.Namespace = config.Name
		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		ctx = messagecontext.NewContext(ctx, messagecontext.NewMessage())

		return ctx, nil
	}

	var operatorFramework *framework.Framework
	{
		frameworkConfig := framework.DefaultConfig()

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

		healthzConfig.K8sClient = k8sClient
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
