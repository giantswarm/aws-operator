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

	SetupCmd.Flags().BoolVar(&remoteCluster, "remote", true, "use remote cluster")
	SetupCmd.Flags().StringVar(&name, "name", "e2e-harness", "CI execution identifier")
}

func runSetup(cmd *cobra.Command, args []string) error {
	logger, err := micrologger.New(micrologger.DefaultConfig())
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

	d := docker.New(logger, e2eHarnessTag, remoteCluster)
	pa := patterns.New(logger)
	w := wait.New(logger, d, pa)
	pCfg := &project.Config{
		Name: projectName,
		Tag:  projectTag,
	}
	fs := afero.NewOsFs()
	pDeps := &project.Dependencies{
		Logger: logger,
		Runner: d,
		Wait:   w,
		Fs:     fs,
	}
	p := project.New(pDeps, pCfg)
	hCfg := harness.Config{
		RemoteCluster: remoteCluster,
	}
	h := harness.New(logger, fs, hCfg)
	c := cluster.New(logger, fs, d, remoteCluster)

	// tasks to run
	bundle := []tasks.Task{
		h.Init,
		h.WriteConfig,
		c.Create,
		p.CommonSetupSteps,
	}

	return tasks.Run(bundle)
}
