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
	TeardownCmd = &cobra.Command{
		Use:   "teardown",
		Short: "teardown e2e tests",
		RunE:  runTeardown,
	}
)

func init() {
	RootCmd.AddCommand(TeardownCmd)
}

func runTeardown(cmd *cobra.Command, args []string) error {
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		return err
	}
	h := harness.New(logger, harness.Config{})
	cfg, err := h.ReadConfig()
	if err != nil {
		return err
	}
	imageTag := GetGitCommit()
	projectName := GetProjectName()

	d := docker.New(logger, imageTag)
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
	c := cluster.New(logger, d, cfg.RemoteCluster)

	bundle := []tasks.Task{
		p.TeardownSteps,
		c.Delete,
	}

	return tasks.Run(bundle)
}
