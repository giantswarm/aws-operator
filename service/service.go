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
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	versionedinfrastructure "github.com/giantswarm/aws-operator/pkg/clientset/versioned"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/collector"
	"github.com/giantswarm/aws-operator/service/controller"
	"github.com/giantswarm/aws-operator/service/locker"
	legacynetwork "github.com/giantswarm/aws-operator/service/network"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service

	bootOnce                sync.Once
	legacyClusterController *controller.Cluster
	legacyDrainerController *controller.Drainer
	operatorCollector       *collector.Set
	statusResourceCollector *statusresource.CollectorSet
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

	cmaClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClientInfra, err := versionedinfrastructure.NewForConfig(restConfig)
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

	var awsConfig aws.Config
	{
		awsConfig = aws.Config{
			AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID),
			AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret),
			Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
			SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
		}
	}

	var kubeLockLocker locker.Interface
	{
		c := locker.KubeLockLockerConfig{
			Logger:     config.Logger,
			RestConfig: restConfig,
		}

		kubeLockLocker, err = locker.NewKubeLockLocker(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyNetworkAllocator legacynetwork.Allocator
	{
		c := legacynetwork.Config{
			Locker: kubeLockLocker,
			Logger: config.Logger,
		}

		legacyNetworkAllocator, err = legacynetwork.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyClusterController *controller.Cluster
	{
		_, ipamNetworkRange, err := net.ParseCIDR(config.Viper.GetString(config.Flag.Service.Installation.Guest.IPAM.Network.CIDR))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := controller.ClusterConfig{
			CMAClient:        cmaClient,
			G8sClient:        g8sClient,
			G8sClientInfra:   g8sClientInfra,
			K8sClient:        k8sClient,
			K8sExtClient:     k8sExtClient,
			Logger:           config.Logger,
			NetworkAllocator: legacyNetworkAllocator,
			APIWhitelist: controller.FrameworkConfigAPIWhitelist{
				Private: controller.FrameworkConfigAPIWhitelistConfig{
					Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.Enabled),
					SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.SubnetList),
				},
				Public: controller.FrameworkConfigAPIWhitelistConfig{
					Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.Enabled),
					SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.SubnetList),
				},
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
			IgnitionPath:              config.Viper.GetString(config.Flag.Service.Guest.Ignition.Path),
			ImagePullProgressDeadline: config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.Kubelet.ImagePullProgressDeadline),
			IncludeTags:               config.Viper.GetBool(config.Flag.Service.AWS.IncludeTags),
			InstallationName:          config.Viper.GetString(config.Flag.Service.Installation.Name),
			IPAMNetworkRange:          *ipamNetworkRange,
			LabelSelector: controller.ClusterConfigLabelSelector{
				Enabled:          config.Viper.GetBool(config.Flag.Service.Feature.LabelSelector.Enabled),
				OverridenVersion: config.Viper.GetString(config.Flag.Service.Test.LabelSelector.Version),
			},
			OIDC: controller.ClusterConfigOIDC{
				ClientID:      config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID),
				IssuerURL:     config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL),
				UsernameClaim: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim),
				GroupsClaim:   config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim),
			},

			PodInfraContainerImage: config.Viper.GetString(config.Flag.Service.AWS.PodInfraContainerImage),
			ProjectName:            project.Name(),
			RegistryDomain:         config.Viper.GetString(config.Flag.Service.RegistryDomain),
			Route53Enabled:         config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
			RouteTables:            config.Viper.GetString(config.Flag.Service.AWS.RouteTables),
			SSOPublicKey:           config.Viper.GetString(config.Flag.Service.Guest.SSH.SSOPublicKey),
			VaultAddress:           config.Viper.GetString(config.Flag.Service.AWS.VaultAddress),
			VPCPeerID:              config.Viper.GetString(config.Flag.Service.AWS.VPCPeerID),
		}

		legacyClusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyDrainerController *controller.Drainer
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
			LabelSelector: controller.DrainerConfigLabelSelector{
				Enabled:          config.Viper.GetBool(config.Flag.Service.Feature.LabelSelector.Enabled),
				OverridenVersion: config.Viper.GetString(config.Flag.Service.Test.LabelSelector.Version),
			},
			ProjectName:    project.Name(),
			Route53Enabled: config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
		}

		legacyDrainerController, err = controller.NewDrainer(c)
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

	var statusResourceCollector *statusresource.CollectorSet
	{
		c := statusresource.CollectorSetConfig{
			Logger:  config.Logger,
			Watcher: g8sClient.ProviderV1alpha1().AWSConfigs("").Watch,
		}

		statusResourceCollector, err = statusresource.NewCollectorSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: []versionbundle.Bundle{project.NewBundle()},
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:                sync.Once{},
		legacyClusterController: legacyClusterController,
		legacyDrainerController: legacyDrainerController,
		operatorCollector:       operatorCollector,
		statusResourceCollector: statusResourceCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)
		go s.statusResourceCollector.Boot(ctx)

		go s.legacyClusterController.Boot(ctx)
		go s.legacyDrainerController.Boot(ctx)
	})
}
