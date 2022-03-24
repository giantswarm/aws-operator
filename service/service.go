// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	corev1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v7/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys/v2"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/images"
	"github.com/giantswarm/aws-operator/service/internal/locker"
	"github.com/giantswarm/aws-operator/service/internal/recorder"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service

	bootOnce                           sync.Once
	clusterController                  *controller.Cluster
	controlPlaneController             *controller.ControlPlane
	controlPlaneDrainerController      *controller.ControlPlaneDrainer
	machineDeploymentController        *controller.MachineDeployment
	machineDeploymentDrainerController *controller.MachineDeploymentDrainer
	terminateUnhealthyNodeController   *controller.TerminateUnhealthyNode
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
				apiv1alpha3.AddToScheme,
				infrastructurev1alpha3.AddToScheme,
				releasev1alpha1.AddToScheme,
				corev1alpha1.AddToScheme,
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
			RoleARN:         config.Viper.GetString(config.Flag.Service.AWS.Role.ARN),
			SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session),
		}
	}

	var certsSearcher *certs.Searcher
	{
		c := certs.Config{
			K8sClient: k8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudtagObject cloudtags.Interface
	{
		c := cloudtags.Config{
			K8sClient: k8sClient,
			Logger:    config.Logger,
		}

		cloudtagObject, err = cloudtags.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var event recorder.Interface
	{
		c := recorder.Config{
			K8sClient: k8sClient,

			Component: fmt.Sprintf("%s-%s", project.Name(), project.Version()),
		}

		event = recorder.New(c)
	}

	var ha hamaster.Interface
	{
		c := hamaster.Config{
			K8sClient: k8sClient,
		}

		ha, err = hamaster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var im images.Interface
	{
		c := images.Config{
			K8sClient: k8sClient,

			RegistryDomain: config.Viper.GetString(config.Flag.Service.Registry.Domain),
		}

		im, err = images.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
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

	var randomKeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: k8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomKeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Cluster
	{
		c := controller.ClusterConfig{
			CloudTags: cloudtagObject,
			Event:     event,
			K8sClient: k8sClient,
			HAMaster:  ha,
			Locker:    kubeLockLocker,
			Logger:    config.Logger,

			AccessLogsExpiration:  config.Viper.GetInt(config.Flag.Service.AWS.S3AccessLogsExpiration),
			AdvancedMonitoringEC2: config.Viper.GetBool(config.Flag.Service.AWS.AdvancedMonitoringEC2),
			APIWhitelist: tccp.ConfigAPIWhitelist{
				Private: tccp.ConfigAPIWhitelistSecurityGroup{
					Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.Enabled),
					SubnetList: strings.Split(config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.SubnetList), ","),
				},
				Public: tccp.ConfigAPIWhitelistSecurityGroup{
					Enabled:    config.Viper.GetBool(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.Enabled),
					SubnetList: strings.Split(config.Viper.GetString(config.Flag.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.SubnetList), ","),
				},
			},
			CalicoCIDR:                 config.Viper.GetInt(config.Flag.Service.Cluster.Calico.CIDR),
			CalicoSubnet:               config.Viper.GetString(config.Flag.Service.Cluster.Calico.Subnet),
			DeleteLoggingBucket:        config.Viper.GetBool(config.Flag.Service.AWS.LoggingBucket.Delete),
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig:              awsConfig,
			IncludeTags:                config.Viper.GetBool(config.Flag.Service.AWS.IncludeTags),
			InstallationName:           config.Viper.GetString(config.Flag.Service.Installation.Name),
			IPAMNetworkRange:           ipamNetworkRange,
			Route53Enabled:             config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
			RouteTables:                config.Viper.GetString(config.Flag.Service.AWS.RouteTables),
		}

		clusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var controlPlaneController *controller.ControlPlane
	{
		c := controller.ControlPlaneConfig{
			CertsSearcher:      certsSearcher,
			CloudTags:          cloudtagObject,
			Event:              event,
			HAMaster:           ha,
			Images:             im,
			K8sClient:          k8sClient,
			Logger:             config.Logger,
			RandomKeysSearcher: randomKeysSearcher,

			CalicoCIDR:                config.Viper.GetInt(config.Flag.Service.Cluster.Calico.CIDR),
			CalicoMTU:                 config.Viper.GetInt(config.Flag.Service.Cluster.Calico.MTU),
			CalicoSubnet:              config.Viper.GetString(config.Flag.Service.Cluster.Calico.Subnet),
			ClusterDomain:             config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.ClusterDomain),
			ClusterIPRange:            config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.API.ClusterIPRange),
			DockerDaemonCIDR:          config.Viper.GetString(config.Flag.Service.Cluster.Docker.Daemon.CIDR),
			DockerhubToken:            config.Viper.GetString(config.Flag.Service.Registry.DockerhubToken),
			ExternalSNAT:              config.Viper.GetBool(config.Flag.Service.AWS.CNI.ExternalSNAT),
			IgnitionPath:              config.Viper.GetString(config.Flag.Service.Guest.Ignition.Path),
			ImagePullProgressDeadline: config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.Kubelet.ImagePullProgressDeadline),
			InstallationName:          config.Viper.GetString(config.Flag.Service.Installation.Name),
			NetworkSetupDockerImage:   config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.NetworkSetup.Docker.Image),
			PodInfraContainerImage:    config.Viper.GetString(config.Flag.Service.AWS.PodInfraContainerImage),
			Route53Enabled:            config.Viper.GetBool(config.Flag.Service.AWS.Route53.Enabled),
			RegistryDomain:            config.Viper.GetString(config.Flag.Service.Registry.Domain),
			RegistryMirrors:           config.Viper.GetStringSlice(config.Flag.Service.Registry.Mirrors),
			SSHUserList:               config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.SSH.UserList),
			SSOPublicKey:              config.Viper.GetString(config.Flag.Service.Guest.SSH.SSOPublicKey),

			HostAWSConfig: awsConfig,
		}

		controlPlaneController, err = controller.NewControlPlane(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var controlPlaneDrainerController *controller.ControlPlaneDrainer
	{
		c := controller.ControlPlaneDrainerConfig{
			Event:     event,
			K8sClient: k8sClient,
			Logger:    config.Logger,

			HostAWSConfig: awsConfig,
		}

		controlPlaneDrainerController, err = controller.NewControlPlaneDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentController *controller.MachineDeployment
	{
		c := controller.MachineDeploymentConfig{
			CertsSearcher:      certsSearcher,
			CloudTags:          cloudtagObject,
			Event:              event,
			HAMaster:           ha,
			Images:             im,
			K8sClient:          k8sClient,
			Locker:             kubeLockLocker,
			Logger:             config.Logger,
			RandomKeysSearcher: randomKeysSearcher,

			AlikeInstances:             config.Viper.GetString(config.Flag.Service.AWS.AlikeInstances),
			CalicoCIDR:                 config.Viper.GetInt(config.Flag.Service.Cluster.Calico.CIDR),
			CalicoMTU:                  config.Viper.GetInt(config.Flag.Service.Cluster.Calico.MTU),
			CalicoSubnet:               config.Viper.GetString(config.Flag.Service.Cluster.Calico.Subnet),
			ClusterDomain:              config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.ClusterDomain),
			ClusterIPRange:             config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.API.ClusterIPRange),
			DockerDaemonCIDR:           config.Viper.GetString(config.Flag.Service.Cluster.Docker.Daemon.CIDR),
			DockerhubToken:             config.Viper.GetString(config.Flag.Service.Registry.DockerhubToken),
			ExternalSNAT:               config.Viper.GetBool(config.Flag.Service.AWS.CNI.ExternalSNAT),
			GuestPrivateSubnetMaskBits: config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits),
			GuestPublicSubnetMaskBits:  config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits),
			GuestSubnetMaskBits:        config.Viper.GetInt(config.Flag.Service.Installation.Guest.IPAM.Network.SubnetMaskBits),
			HostAWSConfig:              awsConfig,
			IgnitionPath:               config.Viper.GetString(config.Flag.Service.Guest.Ignition.Path),
			ImagePullProgressDeadline:  config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.Kubelet.ImagePullProgressDeadline),
			InstallationName:           config.Viper.GetString(config.Flag.Service.Installation.Name),
			IPAMNetworkRange:           ipamNetworkRange,
			NetworkSetupDockerImage:    config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.NetworkSetup.Docker.Image),
			PodInfraContainerImage:     config.Viper.GetString(config.Flag.Service.AWS.PodInfraContainerImage),
			RegistryDomain:             config.Viper.GetString(config.Flag.Service.Registry.Domain),
			RegistryMirrors:            config.Viper.GetStringSlice(config.Flag.Service.Registry.Mirrors),
			RouteTables:                config.Viper.GetString(config.Flag.Service.AWS.RouteTables),
			SSHUserList:                config.Viper.GetString(config.Flag.Service.Cluster.Kubernetes.SSH.UserList),
			SSOPublicKey:               config.Viper.GetString(config.Flag.Service.Guest.SSH.SSOPublicKey),
		}

		machineDeploymentController, err = controller.NewMachineDeployment(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentDrainerController *controller.MachineDeploymentDrainer
	{
		c := controller.MachineDeploymentDrainerConfig{
			Event:     event,
			K8sClient: k8sClient,
			Logger:    config.Logger,

			HostAWSConfig: awsConfig,
		}

		machineDeploymentDrainerController, err = controller.NewMachineDeploymentDrainer(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var terminateUnhealthyNodeController *controller.TerminateUnhealthyNode
	{
		c := controller.TerminateUnhealthyNodeConfig{
			K8sClient: k8sClient,
			Locker:    kubeLockLocker,
			Logger:    config.Logger,

			HostAWSConfig: awsConfig,
		}

		terminateUnhealthyNodeController, err = controller.NewTerminateUnhealthyNode(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:                           sync.Once{},
		clusterController:                  clusterController,
		controlPlaneController:             controlPlaneController,
		controlPlaneDrainerController:      controlPlaneDrainerController,
		machineDeploymentController:        machineDeploymentController,
		machineDeploymentDrainerController: machineDeploymentDrainerController,
		terminateUnhealthyNodeController:   terminateUnhealthyNodeController,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.clusterController.Boot(ctx)
		go s.controlPlaneController.Boot(ctx)
		go s.controlPlaneDrainerController.Boot(ctx)
		go s.machineDeploymentController.Boot(ctx)
		go s.machineDeploymentDrainerController.Boot(ctx)
		go s.terminateUnhealthyNodeController.Boot(ctx)
	})
}
