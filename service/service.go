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
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/service/collector"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi"
	"github.com/giantswarm/aws-operator/service/controller/legacy"
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

	bootOnce                    sync.Once
	clusterapiClusterController *clusterapi.Cluster
	clusterapiDrainerController *clusterapi.Drainer
	legacyClusterController     *legacy.Cluster
	legacyDrainerController     *legacy.Drainer
	operatorCollector           *collector.Set
	statusResourceCollector     *statusresource.Collector
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

			Address:    config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
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

	clusterAPIClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
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

	var clusterapiClusterController *clusterapi.Cluster
	{
		_, ipamNetworkRange, err := net.ParseCIDR(config.Viper.GetString(config.Flag.Service.Installation.Guest.IPAM.Network.CIDR))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := clusterapi.ClusterConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			APIWhitelist: clusterapi.FrameworkConfigAPIWhitelistConfig{
				Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Enabled),
				SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.SubnetList),
			},
			AccessLogsExpiration:  config.Viper.GetInt(config.Flag.Service.AWS.S3AccessLogsExpiration),
			AdvancedMonitoringEC2: config.Viper.GetBool(config.Flag.Service.AWS.AdvancedMonitoringEC2),
			DeleteLoggingBucket:   config.Viper.GetBool(config.Flag.Service.AWS.LoggingBucket.Delete),
			EncrypterBackend:      config.Viper.GetString(config.Flag.Service.AWS.Encrypter),
			GuestAWSConfig: clusterapi.ClusterConfigAWSConfig{
				AccessKeyID:       config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret:   config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				AvailabilityZones: config.Viper.GetStringSlice(config.Flag.Service.AWS.AvailabilityZones),
				SessionToken:      config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:            config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig: clusterapi.ClusterConfigAWSConfig{
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
			OIDC: clusterapi.ClusterConfigOIDC{
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

		clusterapiClusterController, err = clusterapi.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterapiDrainerController *clusterapi.Drainer
	{
		c := clusterapi.DrainerConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			GuestAWSConfig: clusterapi.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			HostAWSConfig: clusterapi.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			ProjectName:    config.ProjectName,
			Route53Enabled: config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
		}

		clusterapiDrainerController, err = clusterapi.NewDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyClusterController *legacy.Cluster
	{
		_, ipamNetworkRange, err := net.ParseCIDR(config.Viper.GetString(config.Flag.Service.Installation.Guest.IPAM.Network.CIDR))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := legacy.ClusterConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			APIWhitelist: legacy.FrameworkConfigAPIWhitelistConfig{
				Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Enabled),
				SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.SubnetList),
			},
			AccessLogsExpiration:  config.Viper.GetInt(config.Flag.Service.AWS.S3AccessLogsExpiration),
			AdvancedMonitoringEC2: config.Viper.GetBool(config.Flag.Service.AWS.AdvancedMonitoringEC2),
			DeleteLoggingBucket:   config.Viper.GetBool(config.Flag.Service.AWS.LoggingBucket.Delete),
			EncrypterBackend:      config.Viper.GetString(config.Flag.Service.AWS.Encrypter),
			GuestAWSConfig: legacy.ClusterConfigAWSConfig{
				AccessKeyID:       config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret:   config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				AvailabilityZones: config.Viper.GetStringSlice(config.Flag.Service.AWS.AvailabilityZones),
				SessionToken:      config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:            config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig: legacy.ClusterConfigAWSConfig{
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
			OIDC: legacy.ClusterConfigOIDC{
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

		legacyClusterController, err = legacy.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyDrainerController *legacy.Drainer
	{
		c := legacy.DrainerConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,
			Logger:       config.Logger,

			GuestAWSConfig: legacy.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			HostAWSConfig: legacy.DrainerConfigAWS{
				AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
				AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
				SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
				Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			},
			ProjectName:    config.ProjectName,
			Route53Enabled: config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
		}

		legacyDrainerController, err = legacy.NewDrainer(c)
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

		bootOnce:                    sync.Once{},
		clusterapiClusterController: clusterapiClusterController,
		clusterapiDrainerController: clusterapiDrainerController,
		legacyClusterController:     legacyClusterController,
		legacyDrainerController:     legacyDrainerController,
		operatorCollector:           operatorCollector,
		statusResourceCollector:     statusResourceCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)
		go s.statusResourceCollector.Boot(ctx)

		//go s.clusterapiClusterController.Boot(ctx)
		//go s.clusterapiDrainerController.Boot(ctx)
		go s.legacyClusterController.Boot(ctx)
		go s.legacyDrainerController.Boot(ctx)
	})
}
