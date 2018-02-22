package cmd

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/e2e-harness/pkg/localkube"
	"github.com/giantswarm/e2e-harness/pkg/tasks"
)

var (
	LocalkubeCmd = &cobra.Command{
		Use:   "localkube",
		Short: "setup localkube",
		RunE:  runLocalkube,
	}
)

func init() {
	RootCmd.AddCommand(LocalkubeCmd)
}

func runLocalkube(cmd *cobra.Command, args []string) error {
	l := localkube.New()

	// tasks to run
	bundle := []tasks.Task{
		l.SetUp,
	}

	return tasks.Run(bundle)
}
