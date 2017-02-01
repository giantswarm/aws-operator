// Package daemon implements the daemon command for any microservice.
package daemon

import (
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"

	microerror "github.com/giantswarm/microkit/error"
	"github.com/giantswarm/microkit/logger"
	"github.com/giantswarm/microkit/server"
)

// Config represents the configuration used to create a new daemon command.
type Config struct {
	// Dependencies.
	Logger        logger.Logger
	ServerFactory func() server.Server
}

// DefaultConfig provides a default configuration to create a new daemon command
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:        nil,
		ServerFactory: nil,
	}
}

// New creates a new daemon command.
func New(config Config) (Command, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.ServerFactory == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "server factory must not be empty")
	}

	newCommand := &command{
		// Internals.
		cobraCommand:  nil,
		logger:        config.Logger,
		serverFactory: config.ServerFactory,
	}

	newCommand.cobraCommand = &cobra.Command{
		Use:   "daemon",
		Short: "Execute the daemon of the microservice.",
		Long:  "Execute the daemon of the microservice.",
		Run:   newCommand.Execute,
	}

	newCommand.cobraCommand.PersistentFlags().StringSliceVar(&Flags.Config.Dirs, "config.dirs", []string{"."}, "List of config file directories.")
	newCommand.cobraCommand.PersistentFlags().StringSliceVar(&Flags.Config.Files, "config.files", []string{"config"}, "List of the config file names. All viper supported extensions can be used.")

	newCommand.cobraCommand.PersistentFlags().StringVar(&Flags.Server.Listen.Address, "server.listen.address", "http://127.0.0.1:8000", "Address used to make the server listen to.")

	return newCommand, nil
}

type command struct {
	// Internals.
	cobraCommand  *cobra.Command
	logger        logger.Logger
	serverFactory func() server.Server
}

func (c *command) CobraCommand() *cobra.Command {
	return c.cobraCommand
}

func (c *command) Execute(cmd *cobra.Command, args []string) {
	// Merge the given command line flags with the given environment variables and
	// the given config file, if any. The merged flags will be applied to the
	// global Flags struct.
	err := MergeFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}

	customServer := c.serverFactory()

	var newServer server.Server
	{
		serverConfig := server.DefaultConfig()

		serverConfig.Endpoints = customServer.Endpoints()
		serverConfig.ErrorEncoder = customServer.ErrorEncoder()
		serverConfig.ListenAddress = Flags.Server.Listen.Address
		serverConfig.Logger = customServer.Logger()
		serverConfig.RequestFuncs = customServer.RequestFuncs()
		serverConfig.Router = customServer.Router()
		serverConfig.ServiceName = customServer.ServiceName()
		serverConfig.TransactionResponder = customServer.TransactionResponder()

		newServer, err = server.New(serverConfig)
		if err != nil {
			panic(err)
		}
		go newServer.Boot()
	}

	// Listen to OS signals.
	listener := make(chan os.Signal, 2)
	signal.Notify(listener, os.Interrupt, os.Kill)

	<-listener

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			newServer.Shutdown()
		}()

		os.Exit(0)
	}()

	<-listener

	os.Exit(0)
}
