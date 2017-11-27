package cmd

import (
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/e2e-harness/pkg/compiler"
	"github.com/giantswarm/e2e-harness/pkg/docker"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/minikube"
	"github.com/giantswarm/e2e-harness/pkg/patterns"
	"github.com/giantswarm/e2e-harness/pkg/project"
	"github.com/giantswarm/e2e-harness/pkg/tasks"
	"github.com/giantswarm/e2e-harness/pkg/wait"
)

var (
	TestCmd = &cobra.Command{
		Use:   "test",
		Short: "execute e2e tests",
		RunE:  runTest,
	}
)

func init() {
	RootCmd.AddCommand(TestCmd)
}

func runTest(cmd *cobra.Command, args []string) error {
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		return err
	}

	projectTag := harness.GetProjectTag()
	projectName := harness.GetProjectName()

	fs := afero.NewOsFs()

	h := harness.New(logger, fs, harness.Config{})
	cfg, err := h.ReadConfig()
	if err != nil {
		return err
	}

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

	comp := compiler.New(logger)

	// tasks to run
	bundle := []tasks.Task{
		comp.CompileTests,
	}

	if !cfg.RemoteCluster {
		// build images for minikube
		m := minikube.New(logger, d, projectTag)

		bundle = append(bundle, comp.CompileMain, m.BuildImages)
	}

	bundle = append(bundle, p.Test)

	return tasks.Run(bundle)
}
