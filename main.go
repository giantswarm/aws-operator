package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/server"
	"github.com/giantswarm/aws-operator/service"
)

var (
	description string     = "The aws-operator handles Kubernetes clusters running on a Kubernetes cluster inside of AWS."
	f           *flag.Flag = flag.New()
	gitCommit   string     = "n/a"
	name        string     = "aws-operator"
	source      string     = "https://github.com/giantswarm/aws-operator"
)

func main() {
	err := mainError()
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
}

func mainError() error {
	var err error

	ctx := context.Background()
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	serverFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{
				Flag:   f,
				Logger: logger,
				Viper:  v,

				Description: description,
				GitCommit:   gitCommit,
				ProjectName: name,
				Source:      source,
			}

			newService, err = service.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", microerror.Mask(err)))
			}

			go newService.Boot(ctx)
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  logger,
				Service: newService,
				Viper:   v,

				ProjectName: name,
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(err)
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

			Description:    description,
			GitCommit:      gitCommit,
			Name:           name,
			Source:         source,
			VersionBundles: service.NewVersionBundles(),
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
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.ID, "", "ID of the AWS access key for the host cluster account. If empty, guest cluster account is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.Secret, "", "Secret of the AWS access key for the host cluster account. If empty, guest cluster account is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.Session, "", "Session token of the AWS access key for the host cluster account. If empty, guest cluster token is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.Region, "", "Region for checking for orphan AWS resources.")

	// TODO(nhlfr): Deprecate these options when cert-operator will be implemented.
	daemonCommand.PersistentFlags().String(f.Service.AWS.PubKeyFile, path.Join(string(os.PathSeparator), ".ssh", "id_rsa.pub"), "Public key to be imported as a keypair in AWS.")

	daemonCommand.PersistentFlags().Bool(f.Service.Guest.Update.Enabled, false, "Whether updates of guest cluster nodes are allowed to be processed upon reconciliation.")

	daemonCommand.PersistentFlags().Bool(f.Service.AWS.LoggingBucket.Delete, false, "Should be logging bucket deleted.")

	daemonCommand.PersistentFlags().Bool(f.Service.AWS.Route53.Enabled, true, "Should Route53 be enabled.")

	daemonCommand.PersistentFlags().String(f.Service.AWS.PodInfraContainerImage, "", "Image to be used for the pause container. If empty, default image from gcr.io/google_containers/pause-amd64 is used.")

	daemonCommand.PersistentFlags().String(f.Service.Installation.Name, "", "Installation name for tagging AWS resources.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.ClientID, "", "OIDC authorization provider ClientID.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.IssuerURL, "", "OIDC authorization provider IssuerURL.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim, "", "OIDC authorization provider UsernameClaim.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim, "", "OIDC authorization provider GroupsClaim.")

	daemonCommand.PersistentFlags().Bool(f.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.Enabled, false, "Enable or disable guest cluster k8s API whitelisting.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Guest.Kubernetes.API.Security.Whitelist.SubnetList, "", "Subnet list for guest cluster k8s API whitelisting.")

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.BearerToken, "", "Token (if needed for Kubernetes authentication).")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Password, "", "Password (if Kubernetes cluster is using basic authentication.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Username, "", "Username (if the Kubernetes cluster is using basic authentication).")
	daemonCommand.PersistentFlags().Int(f.Service.AWS.S3AccessLogsExpiration, 365, "S3 access logs expiration policy.")
	daemonCommand.PersistentFlags().Bool(f.Service.AWS.AdvancedMonitoringEC2, false, "Advanced EC2 monitoring.")

	newCommand.CobraCommand().Execute()

	return nil
}
