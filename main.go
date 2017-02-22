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
		Region string
	}
	Kubernetes struct {
		APIServer   string
		Username    string
		Password    string
		BearerToken string
		TLS         struct {
			Authority struct {
				Certificate     string
				CertificateFile string
			}
			Client struct {
				Certificate     string
				CertificateFile string
				Key             string
				KeyFile         string
			}
		}
		Insecure bool
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
				Region:          Flags.Aws.Region,
			}
			k8sTlsClientConfig := k8sclient.TLSClientConfig{
				CertFile: Flags.Kubernetes.TLS.Client.CertificateFile,
				KeyFile:  Flags.Kubernetes.TLS.Client.KeyFile,
				CAFile:   Flags.Kubernetes.TLS.Authority.Certificate,
				CertData: Flags.Kubernetes.TLS.Client.Certificate,
				KeyData:  Flags.Kubernetes.TLS.Client.Key,
				CAData:   Flags.Kubernetes.TLS.Authority.Certificate,
			}
			serviceConfig.K8sConfig = k8sclient.Config{
				Host:            Flags.Kubernetes.APIServer,
				Username:        Flags.Kubernetes.Username,
				Password:        Flags.Kubernetes.Password,
				BearerToken:     Flags.Kubernetes.BearerToken,
				Insecure:        Flags.Kubernetes.Insecure,
				TLSClientConfig: k8sTlsClientConfig,
			}

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
	daemonCommand.PersistentFlags().StringVar(&Flags.Aws.Region, "aws.region", "", "Region in EC2 service")

	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.APIServer, "kubernetes.apiserver", "http://127.0.0.1:8080", "Address and port of Giantnetes API server")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.Username, "kubernetes.username", "", "Username (if the Kubernetes cluster is using basic authentication)")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.Password, "kubernetes.password", "", "Password (if Kubernetes cluster is using basic authentication")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.BearerToken, "kubernetes.token", "", "Token (if needed for Kubernetes authentication)")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.TLS.Authority.Certificate, "kubernetes.tls.authority.certificate", "", "TLS Authority certificate")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.TLS.Authority.CertificateFile, "kubernetes.tls.authority.certificatefile", "", "TLS Authority certificate file")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.TLS.Client.Certificate, "kubernetes.tls.client.certificate", "", "TLS Client certificate")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.TLS.Client.CertificateFile, "kubernetes.tls.client.certificatefile", "", "TLS Client certificate file")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.TLS.Client.Key, "kubernetes.tls.client.key", "", "TLS Client key")
	daemonCommand.PersistentFlags().StringVar(&Flags.Kubernetes.TLS.Client.KeyFile, "kubernetes.tls.client.keyfile", "", "TLS Client key file")
	daemonCommand.PersistentFlags().BoolVar(&Flags.Kubernetes.Insecure, "kubernetes.insecure", false, "Insecure SSL connection")

	newCommand.CobraCommand().Execute()
}
