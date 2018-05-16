// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/service/alerter"
	"github.com/giantswarm/aws-operator/service/collector"
	"github.com/giantswarm/aws-operator/service/controller"
	"github.com/giantswarm/aws-operator/service/healthz"
)

const (
	RedactedString = "[REDACTED]"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	ProjectName string
	Source      string
}

type Service struct {
	Alerter *alerter.Service
	Healthz *healthz.Service
	Version *version.Service

	metricsCollector  *collector.Collector
	bootOnce          sync.Once
	clusterController *controller.Cluster
	drainerController *controller.Drainer
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:   config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster: config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			TLS: k8srestconfig.TLSClientConfig{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusterController *controller.Cluster
	{
		c := controller.ClusterConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			APIWhitelist: controller.FrameworkConfigAPIWhitelistConfig{
				Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Enabled),
				SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.SubnetList),
			},
			AccessLogsExpiration:  config.Viper.GetInt(config.Flag.Service.AWS.S3AccessLogsExpiration),
			AdvancedMonitoringEC2: config.Viper.GetBool(config.Flag.Service.AWS.AdvancedMonitoringEC2),
			DeleteLoggingBucket:   config.Viper.GetBool(config.Flag.Service.AWS.LoggingBucket.Delete),
			GuestAWSConfig: controller.ClusterConfigAWSConfig{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			GuestUpdateEnabled: config.Viper.GetBool(config.Flag.Service.Guest.Update.Enabled),
			HostAWSConfig: controller.ClusterConfigAWSConfig{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			InstallationName: config.Viper.GetString(config.Flag.Service.Installation.Name),
			OIDC: controller.ClusterConfigOIDC{
				ClientID:      config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID),
				IssuerURL:     config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL),
				UsernameClaim: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim),
				GroupsClaim:   config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim),
			},

			ProjectName: config.ProjectName,
			PubKeyFile:  config.Viper.GetString(config.Flag.Service.AWS.PubKeyFile),
		}

		clusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var drainerController *controller.Drainer
	{
		c := controller.DrainerConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			AWS: controller.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			GuestUpdateEnabled: config.Viper.GetBool(config.Flag.Service.Guest.Update.Enabled),
			ProjectName:        config.ProjectName,
		}

		drainerController, err = controller.NewDrainer(c)
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

	var alerterService *alerter.Service
	{
		// Set the region, in the operator this comes from the cluster object.
		awsConfig.Region = config.Viper.GetString(config.Flag.Service.AWS.Region)

		alerterConfig := alerter.DefaultConfig()
		alerterConfig.AwsConfig = awsConfig
		alerterConfig.InstallationName = config.Viper.GetString(config.Flag.Service.Installation.Name)
		alerterConfig.G8sClient = g8sClient
		alerterConfig.Logger = config.Logger

		alerterService, err = alerter.New(alerterConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var metricsCollector *collector.Collector
	{
		c := collector.Config{
			Logger: config.Logger,

			AwsConfig:        awsConfig,
			InstallationName: config.Viper.GetString(config.Flag.Service.Installation.Name),
		}

		metricsCollector, err = collector.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	fmt.Printf("%#v\n", metricsCollector)

	var healthzService *healthz.Service
	{
		c := healthz.Config{
			AwsConfig: awsConfig,
			Logger:    config.Logger,
		}
		healthzService, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Description = config.Description
		versionConfig.GitCommit = config.GitCommit
		versionConfig.Name = config.ProjectName
		versionConfig.Source = config.Source
		versionConfig.VersionBundles = NewVersionBundles()

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Alerter: alerterService,
		Healthz: healthzService,
		Version: versionService,

		metricsCollector:  metricsCollector,
		bootOnce:          sync.Once{},
		clusterController: clusterController,
		drainerController: drainerController,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		s.Alerter.StartAlerts()

		fmt.Printf("%#v\n", s.metricsCollector)
		prometheus.MustRegister(s.metricsCollector)

		go s.clusterController.Boot()
		go s.drainerController.Boot()
	})
}
