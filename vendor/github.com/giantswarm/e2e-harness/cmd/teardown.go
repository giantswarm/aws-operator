package cmd

import (
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/e2e-harness/pkg/cluster"
	"github.com/giantswarm/e2e-harness/pkg/docker"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/patterns"
	"github.com/giantswarm/e2e-harness/pkg/project"
	"github.com/giantswarm/e2e-harness/pkg/tasks"
	"github.com/giantswarm/e2e-harness/pkg/wait"
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

	fs := afero.NewOsFs()

	h := harness.New(logger, fs, harness.Config{})
	cfg, err := h.ReadConfig()
	if err != nil {
		return err
	}
	projectTag := harness.GetProjectTag()
	projectName := harness.GetProjectName()

	// use latest tag for consumer projects (not dog-fooding e2e-harness)
	e2eHarnessTag := projectTag
	if projectName != "e2e-harness" {
		e2eHarnessTag = "latest"
	}

	d := docker.New(logger, e2eHarnessTag, cfg.RemoteCluster)
	pa := patterns.New(logger)
	w := wait.New(logger, d, pa)
	pCfg := &project.Config{
		Name: projectName,
		Tag:  projectTag,
	}
	pDeps := &project.Dependencies{
		Logger: logger,
		Runner: d,
		Wait:   w,
		Fs:     fs,
	}
	p := project.New(pDeps, pCfg)
	c := cluster.New(logger, fs, d, cfg.RemoteCluster)

	bundle := []tasks.Task{}

	if cfg.RemoteCluster {
		bundle = append(bundle, c.Delete)
	} else {
		bundle = append(bundle, p.CommonTearDownSteps)
	}

	return tasks.Run(bundle)
}
