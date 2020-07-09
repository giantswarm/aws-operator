package main

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/server"
	"github.com/giantswarm/aws-operator/service"
)

var (
	f *flag.Flag = flag.New()
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func mainE(ctx context.Context) error {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is sorted out.
	serverFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{
				Logger: logger,

				Flag:  f,
				Viper: v,
			}

			newService, err = service.New(c)
			if err != nil {
				panic(microerror.JSON(err))
			}

			go newService.Boot(ctx)
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  logger,
				Service: newService,

				Viper: v,
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        logger,
			ServerFactory: serverFactory,

			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.AWS.AccessKey.ID, "", "ID of the AWS access key for the account to create guest clusters in.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.AccessKey.Secret, "", "Secret of the AWS access key for the  account to create guest clusters in.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.AccessKey.Session, "", "Session token of the AWS access key for the  account to create guest clusters in. (Can be empty)")
	daemonCommand.PersistentFlags().StringSlice(f.Service.AWS.AvailabilityZones, []string{}, "Availability zones as a slice.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.ID, "", "ID of the AWS access key for the host cluster account. If empty, guest cluster account is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.Secret, "", "Secret of the AWS access key for the host cluster account. If empty, guest cluster account is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.Session, "", "Session token of the AWS access key for the host cluster account. If empty, guest cluster token is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.Region, "", "Region for checking for orphaned AWS resources.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.RouteTables, "", "Names of the public route tables in control plane separated by commas, required for accessing public ELBs from tenant nodes.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.VaultAddress, "", "Server address for Vault encryption.")
	daemonCommand.PersistentFlags().Bool(f.Service.AWS.AdvancedMonitoringEC2, false, "Advanced EC2 monitoring.")
	daemonCommand.PersistentFlags().Bool(f.Service.AWS.LoggingBucket.Delete, false, "Should be logging bucket deleted.")
	daemonCommand.PersistentFlags().Bool(f.Service.AWS.Route53.Enabled, true, "Should Route 53 be enabled.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.PodInfraContainerImage, "", "Image to be used for the pause container. If empty, default image from gcr.io/google_containers/pause-amd64 is used.")
	daemonCommand.PersistentFlags().Bool(f.Service.AWS.IncludeTags, true, "Should resource tags be included (especially for restricted regions, like S3 buckets in China regions).")
	daemonCommand.PersistentFlags().Int(f.Service.AWS.S3AccessLogsExpiration, 365, "S3 access logs expiration policy.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.TrustedAdvisor.Enabled, "", "Whether trusted advisor metrics collection is enabled.")
	daemonCommand.PersistentFlags().Bool(f.Service.AWS.CNI.ExternalSNAT, false, "Whether External SNAT for the AWS CNI is enabled.")

	daemonCommand.PersistentFlags().Int(f.Service.Cluster.Calico.CIDR, 0, "Calico CIDR of guest clusters.")
	daemonCommand.PersistentFlags().Int(f.Service.Cluster.Calico.MTU, 0, "Calico MTU of guest clusters.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Calico.Subnet, "", "Calico subnet of guest clusters.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Docker.Daemon.CIDR, "", "CIDR of the Docker daemon bridge configured in guest clusters.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Kubernetes.API.ClusterIPRange, "", "Service IP range within guest clusters.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Kubernetes.ClusterDomain, "", "Internal Kubernetes domain.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Kubernetes.Kubelet.ImagePullProgressDeadline, "1m", "If no progress is made before this deadline image pulling is cancelled.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Kubernetes.NetworkSetup.Docker.Image, "", "Full docker image of networksetup.")
	daemonCommand.PersistentFlags().String(f.Service.Cluster.Kubernetes.SSH.UserList, "", "Comma separated list of ssh users and their public key in format `username:publickey`, being installed in the guest cluster nodes.")

	daemonCommand.PersistentFlags().String(f.Service.Guest.Ignition.Path, "/opt/ignition", "Default path for the ignition base directory.")
	daemonCommand.PersistentFlags().String(f.Service.Guest.SSH.SSOPublicKey, "", "Public key for trusted SSO CA.")

	daemonCommand.PersistentFlags().String(f.Service.Installation.Name, "", "Installation name for tagging AWS resources.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.IPAM.Network.CIDR, "", "Guest cluster network segment from which IPAM allocates subnets.")
	daemonCommand.PersistentFlags().Int(f.Service.Installation.Guest.IPAM.Network.SubnetMaskBits, 24, "Number of bits in guest cluster subnet network mask.")
	daemonCommand.PersistentFlags().Int(f.Service.Installation.Guest.IPAM.Network.PrivateSubnetMaskBits, 25, "Number of bits in guest cluster private subnet network mask. This must be smaller than SubnetMaskBits.")
	daemonCommand.PersistentFlags().Int(f.Service.Installation.Guest.IPAM.Network.PublicSubnetMaskBits, 25, "Number of bits in guest cluster public subnet network mask. This must be smaller than SubnetMaskBits.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID, "", "OIDC authorization provider ClientID.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL, "", "OIDC authorization provider IssuerURL.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim, "", "OIDC authorization provider UsernameClaim.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim, "", "OIDC authorization provider GroupsClaim.")
	daemonCommand.PersistentFlags().Bool(f.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.Enabled, false, "Enable or disable guest cluster k8s private API whitelisting.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Private.SubnetList, "", "Subnet list for guest cluster k8s private API whitelisting.")
	daemonCommand.PersistentFlags().Bool(f.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.Enabled, false, "Enable or disable guest cluster k8s public API whitelisting.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Public.SubnetList, "", "Subnet list for guest cluster k8s public API whitelisting.")

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfig, "", "KubeConfig used to connect to Kubernetes. When empty other settings are used.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	daemonCommand.PersistentFlags().Bool(f.Service.NodeAutoRepair, false, "Enable Node Auto repair feature.")
	daemonCommand.PersistentFlags().String(f.Service.Registry.Domain, "docker.io", "Image registry domain.")
	daemonCommand.PersistentFlags().StringSlice(f.Service.Registry.Mirrors, []string{}, `Image registry mirror domains. Can be set only if registry domain is "docker.io".`)

	err = newCommand.CobraCommand().Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
