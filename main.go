package main

import (
	"os"
	"path"

	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/microkit/transaction"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
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
	var err error

	// Create a new logger which is used by all packages.
	var newLogger micrologger.Logger
	{
		loggerConfig := micrologger.DefaultConfig()
		loggerConfig.IOWriter = os.Stdout
		newLogger, err = micrologger.New(loggerConfig)
		if err != nil {
			panic(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			serviceConfig := service.DefaultConfig()

			serviceConfig.Flag = f
			serviceConfig.Logger = newLogger
			serviceConfig.Viper = v

			serviceConfig.Description = description
			serviceConfig.GitCommit = gitCommit
			serviceConfig.Name = name
			serviceConfig.Source = source

			newService, err = service.New(serviceConfig)
			if err != nil {
				panic(err)
			}
			go newService.Boot()
		}

		var storage microstorage.Storage
		{
			storage, err = memory.New(memory.DefaultConfig())
			if err != nil {
				panic(err)
			}
		}

		var transactionResponder transaction.Responder
		{
			c := transaction.DefaultResponderConfig()
			c.Logger = newLogger
			c.Storage = storage

			transactionResponder, err = transaction.NewResponder(c)
			if err != nil {
				panic(err)
			}
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			serverConfig := server.DefaultConfig()

			serverConfig.MicroServerConfig.Logger = newLogger
			serverConfig.MicroServerConfig.TransactionResponder = transactionResponder
			serverConfig.MicroServerConfig.ServiceName = name
			serverConfig.MicroServerConfig.Viper = v
			serverConfig.Service = newService

			newServer, err = server.New(serverConfig)
			if err != nil {
				panic(err)
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		commandConfig := command.DefaultConfig()

		commandConfig.Logger = newLogger
		commandConfig.ServerFactory = newServerFactory

		commandConfig.Description = description
		commandConfig.GitCommit = gitCommit
		commandConfig.Name = name
		commandConfig.Source = source

		newCommand, err = command.New(commandConfig)
		if err != nil {
			panic(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.AWS.AccessKey.ID, "", "ID of the AWS access key for the account to create guest clusters in.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.AccessKey.Secret, "", "Secret of the AWS access key for the  account to create guest clusters in.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.AccessKey.Session, "", "Session token of the AWS access key for the  account to create guest clusters in. (Can be empty)")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.ID, "", "ID of the AWS access key for the host cluster account. If empty, guest cluster account is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.Secret, "", "Secret of the AWS access key for the host cluster account. If empty, guest cluster account is used.")
	daemonCommand.PersistentFlags().String(f.Service.AWS.HostAccessKey.Session, "", "Session token of the AWS access key for the host cluster account. If empty, guest cluster token is used.")

	// TODO(nhlfr): Deprecate these options when cert-operator will be implemented.
	daemonCommand.PersistentFlags().String(f.Service.AWS.PubKeyFile, path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub"), "Public key to be imported as a keypair in AWS.")

	daemonCommand.PersistentFlags().String(f.Service.Installation.Name, "", "Installation name for tagging AWS resources.")

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.BearerToken, "", "Token (if needed for Kubernetes authentication).")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Password, "", "Password (if Kubernetes cluster is using basic authentication.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Username, "", "Username (if the Kubernetes cluster is using basic authentication).")

	newCommand.CobraCommand().Execute()
}
