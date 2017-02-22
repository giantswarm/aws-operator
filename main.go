package main

import (
	"os"

	"github.com/giantswarm/microkit/command"
	"github.com/giantswarm/microkit/logger"
	microserver "github.com/giantswarm/microkit/server"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	k8sclient "github.com/giantswarm/aws-operator/client/k8s"
	"github.com/giantswarm/aws-operator/server"
	"github.com/giantswarm/aws-operator/service"
)

var (
	description string = "The aws-operator handles Kubernetes clusters running on a Kubernetes cluster inside of AWS."
	gitCommit   string = "n/a"
	name        string = "aws-operator"
	source      string = "https://github.com/giantswarm/aws-operator"
)

// Flags is the global flag structure used to apply certain configuration to it.
// This is used to bundle configuration for the command, server and service
// initialisation.
var Flags = struct {
	Aws struct {
		AccessKey struct {
			ID     string
			Secret string
		}
		CertsDir              string
		CloudconfigMasterPath string
		CloudconfigWorkerPath string
	}
	Kubernetes struct {
		APIServer   string
		Username    string
		Password    string
		BearerToken string
		Insecure    bool
	}
}{}

func main() {
	var err error

	// Create a new logger which is used by all packages.
	var newLogger logger.Logger
	{
		loggerConfig := logger.DefaultConfig()
		loggerConfig.IOWriter = os.Stdout
		newLogger, err = logger.New(loggerConfig)
		if err != nil {
			panic(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	newServerFactory := func() microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			serviceConfig := service.DefaultConfig()

			serviceConfig.Logger = newLogger

			serviceConfig.AwsConfig = awsclient.Config{
				AccessKeyID:     Flags.Aws.AccessKey.ID,
				AccessKeySecret: Flags.Aws.AccessKey.Secret,
			}
			serviceConfig.K8sConfig = k8sclient.Config{
				Host:        Flags.Kubernetes.APIServer,
				Username:    Flags.Kubernetes.Username,
				Password:    Flags.Kubernetes.Password,
				BearerToken: Flags.Kubernetes.BearerToken,
				Insecure:    Flags.Kubernetes.Insecure,
			}

			serviceConfig.Description = description
			serviceConfig.GitCommit = gitCommit
			serviceConfig.Name = name
			serviceConfig.Source = source
			serviceConfig.CertsDir = Flags.Aws.CertsDir
			serviceConfig.CloudconfigMasterPath = Flags.Aws.CloudconfigMasterPath
			serviceConfig.CloudconfigWorkerPath = Flags.Aws.CloudconfigWorkerPath

			newService, err = service.New(serviceConfig)
			if err != nil {
				panic(err)
			}
			go newService.Boot()
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			serverConfig := server.DefaultConfig()

			serverConfig.Logger = newLogger
			serverConfig.Service = newService

			serverConfig.ServiceName = name

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

	daemonCommand.PersistentFlags().StringVar(&Flags.Aws.AccessKey.ID, "aws.accesskey.id", "", "ID of the AWS access key")
	daemonCommand.PersistentFlags().StringVar(&Flags.Aws.AccessKey.Secret, "aws.accesskey.secret", "", "Secret of the AWS access key")
	// TODO move this to the TPR
	daemonCommand.PersistentFlags().StringVar(&Flags.Aws.CertsDir, "aws.certsdir", "", "Certificates to be placed in /etc/kubernetes/ssl")
	daemonCommand.PersistentFlags().StringVar(&Flags.Aws.CloudconfigMasterPath, "aws.cloudconfigMasterPath", "", "Path to the master node cloudconfig template")
	daemonCommand.PersistentFlags().StringVar(&Flags.Aws.CloudconfigWorkerPath, "aws.cloudconfigWorkerPath", "", "Path to the master node cloudconfig template")

	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.APIServer, "kubernetes.apiserver", "http://127.0.0.1:8080", "Address and port of Giantnetes API server")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.Username, "kubernetes.username", "", "Username (if the Kubernetes cluster is using basic authentication)")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.Password, "kubernetes.password", "", "Password (if Kubernetes cluster is using basic authentication")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.BearerToken, "kubernetes.token", "", "Token (if needed for Kubernetes authentication)")
	daemonCommand.PersistentFlags().BoolVar(&Flags.Kubernetes.Insecure, "kubernetes.insecure", false, "Insecure SSL connection")

	newCommand.CobraCommand().Execute()
}
