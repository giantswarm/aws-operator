// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"net"
	"sync"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/collector"
	"github.com/giantswarm/aws-operator/service/controller"
	"github.com/giantswarm/aws-operator/service/internal/locker"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service

	bootOnce                              sync.Once
	clusterapiClusterController           *controller.Cluster
	clusterapiDrainerController           *controller.Drainer
	clusterapiMachineDeploymentController *controller.MachineDeployment
	operatorCollector                     *collector.Set
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

	var k8sClient *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha2.AddToScheme,
				infrastructurev1alpha2.AddToScheme,
			},
			Logger:     config.Logger,
			RestConfig: restConfig,
		}
		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
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

	var ipamNetworkRange net.IPNet
	{
		_, ipnet, err := net.ParseCIDR(config.Viper.GetString(config.Flag.Service.Installation.Guest.IPAM.Network.CIDR))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		ipamNetworkRange = *ipnet
	}

	var clusterapiClusterController *controller.Cluster
	{

		c := controller.ClusterConfig{
			K8sClient: k8sClient,
			Locker:    kubeLockLocker,
			Logger:    config.Logger,

			AccessLogsExpiration:  config.Viper.GetInt(config.Flag.Service.AWS.S3AccessLogsExpiration),
			AdvancedMonitoringEC2: config.Viper.GetBool(config.Flag.Service.AWS.AdvancedMonitoringEC2),
			APIWhitelist: controller.ClusterConfigAPIWhitelist{
				Private: controller.ClusterConfigAPIWhitelistConfig{
					Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.Enabled),
					SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.SubnetList),
				},
				Public: controller.ClusterConfigAPIWhitelistConfig{
					Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.Enabled),
					SubnetList: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.SubnetList)},
			},
			CalicoCIDR:                 config.Viper.GetInt(config.Flag.Service.Cluster.Calico.CIDR),
			CalicoMTU:                  config.Viper.GetInt(config.Flag.Service.Cluster.Calico.MTU),
			CalicoSubnet:               config.Viper.GetString(config.Flag.Service.Cluster.Calico.Subnet),
			ClusterIPRange:             config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.API.ClusterIPRange),
			DeleteLoggingBucket:        config.Viper.GetBool(config.Flag.Service.AWS.LoggingBucket.Delete),
			DockerDaemonCIDR:           config.Viper.GetString(config.Flag.Service.Cluster.Docker.Daemon.CIDR),
			EncrypterBackend:           config.Viper.GetString(config.Flag.Service.AWS.Encrypter),
			GuestAvailabilityZones:     config.Viper.GetStringSlice(config.Flag.Service.AWS.AvailabilityZones),
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig:              awsConfig,
			IgnitionPath:               config.Viper.GetString(config.Flag.Service.Guest.Ignition.Path),
			ImagePullProgressDeadline:  config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.Kubelet.ImagePullProgressDeadline),
			IncludeTags:                config.Viper.GetBool(config.Flag.Service.AWS.IncludeTags),
			InstallationName:           config.Viper.GetString(config.Flag.Service.Installation.Name),
			IPAMNetworkRange:           ipamNetworkRange,
			LabelSelector: controller.ClusterConfigLabelSelector{
				Enabled:          config.Viper.GetBool(config.Flag.Service.Feature.LabelSelector.Enabled),
				OverridenVersion: config.Viper.GetString(config.Flag.Service.Test.LabelSelector.Version),
			},
			NetworkSetupDockerImage: config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.NetworkSetup.Docker.Image),
			OIDC: controller.ClusterConfigOIDC{
				ClientID:      config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID),
				IssuerURL:     config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL),
				UsernameClaim: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim),
				GroupsClaim:   config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim),
			},
			PodInfraContainerImage: config.Viper.GetString(config.Flag.Service.AWS.PodInfraContainerImage),
			RegistryDomain:         config.Viper.GetString(config.Flag.Service.RegistryDomain),
			Route53Enabled:         config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
			RouteTables:            config.Viper.GetString(config.Flag.Service.AWS.RouteTables),
			SSHUserList:            config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.SSH.UserList),
			SSOPublicKey:           config.Viper.GetString(config.Flag.Service.Guest.SSH.SSOPublicKey),
			VaultAddress:           config.Viper.GetString(config.Flag.Service.AWS.VaultAddress),
		}
		clusterapiClusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterapiDrainerController *controller.Drainer
	{
		c := controller.DrainerConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,

			HostAWSConfig: awsConfig,
			LabelSelector: controller.DrainerConfigLabelSelector{
				Enabled:          config.Viper.GetBool(config.Flag.Service.Feature.LabelSelector.Enabled),
				OverridenVersion: config.Viper.GetString(config.Flag.Service.Test.LabelSelector.Version),
			},
			Route53Enabled: config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
		}

		clusterapiDrainerController, err = controller.NewDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterapiMachineDeploymentController *controller.MachineDeployment
	{
		c := controller.MachineDeploymentConfig{
			K8sClient: k8sClient,
			Locker:    kubeLockLocker,
			Logger:    config.Logger,

			CalicoCIDR:                 config.Viper.GetInt(config.Flag.Service.Cluster.Calico.CIDR),
			CalicoMTU:                  config.Viper.GetInt(config.Flag.Service.Cluster.Calico.MTU),
			CalicoSubnet:               config.Viper.GetString(config.Flag.Service.Cluster.Calico.Subnet),
			ClusterIPRange:             config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.API.ClusterIPRange),
			DeleteLoggingBucket:        config.Viper.GetBool(config.Flag.Service.AWS.LoggingBucket.Delete),
			DockerDaemonCIDR:           config.Viper.GetString(config.Flag.Service.Cluster.Docker.Daemon.CIDR),
			EncrypterBackend:           config.Viper.GetString(config.Flag.Service.AWS.Encrypter),
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig:              awsConfig,
			IgnitionPath:               config.Viper.GetString(config.Flag.Service.Guest.Ignition.Path),
			ImagePullProgressDeadline:  config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.Kubelet.ImagePullProgressDeadline),
			InstallationName:           config.Viper.GetString(config.Flag.Service.Installation.Name),
			IPAMNetworkRange:           ipamNetworkRange,
			LabelSelector: controller.MachineDeploymentConfigLabelSelector{
				Enabled:          config.Viper.GetBool(config.Flag.Service.Feature.LabelSelector.Enabled),
				OverridenVersion: config.Viper.GetString(config.Flag.Service.Test.LabelSelector.Version),
			},
			NetworkSetupDockerImage: config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.NetworkSetup.Docker.Image),
			OIDC: controller.ClusterConfigOIDC{
				ClientID:      config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID),
				IssuerURL:     config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL),
				UsernameClaim: config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim),
				GroupsClaim:   config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim),
			},
			PodInfraContainerImage: config.Viper.GetString(config.Flag.Service.AWS.PodInfraContainerImage),
			RegistryDomain:         config.Viper.GetString(config.Flag.Service.RegistryDomain),
			Route53Enabled:         config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
			RouteTables:            config.Viper.GetString(config.Flag.Service.AWS.RouteTables),
			SSHUserList:            config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.SSH.UserList),
			SSOPublicKey:           config.Viper.GetString(config.Flag.Service.Guest.SSH.SSOPublicKey),
			VaultAddress:           config.Viper.GetString(config.Flag.Service.AWS.VaultAddress),
		}

		clusterapiMachineDeploymentController, err = controller.NewMachineDeployment(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			G8sClient: k8sClient.G8sClient(),
			K8sClient: k8sClient.K8sClient(),
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

	var versionService *version.Service
	{
		c := version.Config{
			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: []versionbundle.Bundle{project.NewVersionBundle()},
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:                              sync.Once{},
		clusterapiClusterController:           clusterapiClusterController,
		clusterapiDrainerController:           clusterapiDrainerController,
		clusterapiMachineDeploymentController: clusterapiMachineDeploymentController,
		operatorCollector:                     operatorCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx)

		go s.clusterapiClusterController.Boot(ctx)
		go s.clusterapiDrainerController.Boot(ctx)
		go s.clusterapiMachineDeploymentController.Boot(ctx)
	})
}
