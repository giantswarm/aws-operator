package cmd

import (
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/initializer"
	"github.com/giantswarm/e2e-harness/pkg/tasks"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	InitCmd = &cobra.Command{
		Use:   "init",
		Short: "initialize project to develop and run k8s e2e tests",
		RunE:  runInit,
	}
)

func init() {
	RootCmd.AddCommand(InitCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		return err
	}

	projectName := harness.GetProjectName()
	fs := afero.NewOsFs()
	i := initializer.New(logger, fs, projectName)

	// tasks to run.
	bundle := []tasks.Task{
		i.CreateLayout,
	}

	return tasks.Run(bundle)
}
