package main

import (
	"os"

	"github.com/giantswarm/microkit/command"
	"github.com/giantswarm/microkit/logger"
	microserver "github.com/giantswarm/microkit/server"

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
	Service struct {
		Operator struct {
			Foo string
		}
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

			serviceConfig.OperatorFoo = Flags.Service.Operator.Foo

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

	daemonCommand.PersistentFlags().StringVar(&Flags.Service.Operator.Foo, "service.operator.foo", "bar", "TODO")

	newCommand.CobraCommand().Execute()
}
