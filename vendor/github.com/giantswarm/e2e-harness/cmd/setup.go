package cmd

import (
	"github.com/giantswarm/e2e-harness/pkg/cluster"
	"github.com/giantswarm/e2e-harness/pkg/docker"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/patterns"
	"github.com/giantswarm/e2e-harness/pkg/project"
	"github.com/giantswarm/e2e-harness/pkg/tasks"
	"github.com/giantswarm/e2e-harness/pkg/wait"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"
)

var (
	SetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "setup e2e tests",
		RunE:  runSetup,
	}
	remoteCluster bool
	name          string
)

func init() {
	RootCmd.AddCommand(SetupCmd)

	SetupCmd.Flags().BoolVar(&remoteCluster, "remote-cluster", true, "use remote cluster")
	SetupCmd.Flags().StringVar(&name, "name", "e2e-harness", "CI execution identifier")
}

func runSetup(cmd *cobra.Command, args []string) error {
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		return err
	}

	gitCommit := GetGitCommit()
	projectName := GetProjectName()

	d := docker.New(logger, gitCommit)
	pa := patterns.New(logger)
	w := wait.New(logger, d, pa)
	pCfg := &project.Config{
		Name:      projectName,
		GitCommit: gitCommit,
	}
	pDeps := &project.Dependencies{
		Logger: logger,
		Runner: d,
		Wait:   w,
	}
	p := project.New(pDeps, pCfg)
	hCfg := harness.Config{
		RemoteCluster: remoteCluster,
	}
	h := harness.New(logger, hCfg)
	c := cluster.New(logger, d, remoteCluster)

	// tasks to run
	bundle := []tasks.Task{
		h.Init,
		c.Create,
		p.CommonSetupSteps,
		p.SetupSteps,
		h.WriteConfig,
	}

	return tasks.Run(bundle)
}
