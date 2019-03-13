// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"net"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/statusresource"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/service/collector"
	"github.com/giantswarm/aws-operator/service/controller"
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
	Version *version.Service

	bootOnce                sync.Once
	clusterController       *controller.Cluster
	drainerController       *controller.Drainer
	operatorCollector       *collector.Set
	statusResourceCollector *statusresource.Collector
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
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
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

	var awsConfig clientaws.Config
	{
		awsConfig = clientaws.Config{
			AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
			AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
			Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
		}
	}

	var clusterController *controller.Cluster
	{
		_, ipamNetworkRange, err := net.ParseCIDR(config.Viper.GetString(config.Flag.Service.Installation.Guest.IPAM.Network.CIDR))
		if err != nil {
			return nil, microerror.Mask(err)
		}

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
			EncrypterBackend:      config.Viper.GetString(config.Flag.Service.AWS.Encrypter),
			GuestAWSConfig: controller.ClusterConfigAWSConfig{
				AccessKeyID:       config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret:   config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				AvailabilityZones: config.Viper.GetStringSlice(config.Flag.Service.AWS.AvailabilityZones),
				SessionToken:      config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:            config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig: controller.ClusterConfigAWSConfig{
				AccessKeyID:       config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
				AccessKeySecret:   config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
				AvailabilityZones: config.Viper.GetStringSlice(config.Flag.Service.AWS.AvailabilityZones),
				SessionToken:      config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
				Region:            config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			IgnitionPath:     config.Viper.GetString(config.Flag.Service.Guest.Ignition.Path),
			IncludeTags:      config.Viper.GetBool(config.Flag.Service.AWS.IncludeTags),
			InstallationName: config.Viper.GetString(config.Flag.Service.Installation.Name),
			IPAMNetworkRange: *ipamNetworkRange,
			OIDC: controller.ClusterConfigOIDC{
				ClientID:      config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID),
				IssuerURL:     config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL),
				UsernameClaim: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim),
				GroupsClaim:   config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim),
			},

			PodInfraContainerImage: config.Viper.GetString(config.Flag.Service.AWS.PodInfraContainerImage),
			ProjectName:            config.ProjectName,
			PubKeyFile:             config.Viper.GetString(config.Flag.Service.AWS.PubKeyFile),
			RegistryDomain:         config.Viper.GetString(config.Flag.Service.RegistryDomain),
			Route53Enabled:         config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
			RouteTables:            config.Viper.GetString(config.Flag.Service.AWS.RouteTables),
			SSOPublicKey:           config.Viper.GetString(config.Flag.Service.Guest.SSH.SSOPublicKey),
			VaultAddress:           config.Viper.GetString(config.Flag.Service.AWS.VaultAddress),
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

			GuestAWSConfig: controller.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			HostAWSConfig: controller.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			ProjectName:    config.ProjectName,
			Route53Enabled: config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
		}

		drainerController, err = controller.NewDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			G8sClient: g8sClient,
			K8sClient: k8sClient,
			Logger:    config.Logger,

			AWSConfig:             awsConfig,
			InstallationName:      config.Viper.GetString(config.Flag.Service.Installation.Name),
			TrustedAdvisorEnabled: config.Viper.GetBool(config.Flag.Service.AWS.TrustedAdvisor.Enabled),
		}

		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResourceCollector *statusresource.Collector
	{
		c := statusresource.CollectorConfig{
			Logger:  config.Logger,
			Watcher: g8sClient.ProviderV1alpha1().AWSConfigs("").Watch,
		}

		statusResourceCollector, err = statusresource.NewCollector(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description:    config.Description,
			GitCommit:      config.GitCommit,
			Name:           config.ProjectName,
			Source:         config.Source,
			VersionBundles: NewVersionBundles(),
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:                sync.Once{},
		clusterController:       clusterController,
		drainerController:       drainerController,
		operatorCollector:       operatorCollector,
		statusResourceCollector: statusResourceCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)
		go s.statusResourceCollector.Boot(ctx)

		go s.clusterController.Boot(ctx)
		go s.drainerController.Boot(ctx)
	})
}
