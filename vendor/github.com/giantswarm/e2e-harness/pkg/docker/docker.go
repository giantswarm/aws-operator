package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/micrologger"
)

type Docker struct {
	logger   micrologger.Logger
	imageTag string
}

func New(logger micrologger.Logger, imageTag string) *Docker {
	return &Docker{
		logger:   logger,
		imageTag: imageTag,
	}
}

// RunPortForward executes a command in the e2e-harness container after
// setting up the port forwarding to the remote cluster, this command
// is meant to be used after that cluster has been initialized
func (d *Docker) RunPortForward(out io.Writer, command string) error {
	args := append([]string{"quay.io/giantswarm/e2e-harness:" + d.imageTag}, "-c",
		fmt.Sprintf("shipyard -action=forward-port && %s", command))

	return d.baseRun(out, "/bin/bash", args)
}

// Run executes a command in the e2e-harness container.
func (d *Docker) Run(out io.Writer, command string) error {
	var args []string
	fields := strings.Fields(command)
	if len(fields) > 1 {
		args = fields[1:]
	}

	args = append([]string{"quay.io/giantswarm/e2e-harness:" + d.imageTag}, args...)

	return d.baseRun(out, fields[0], args)
}

func (d *Docker) baseRun(out io.Writer, entrypoint string, args []string) error {
	dir, err := harness.BaseDir()
	if err != nil {
		return err
	}

	baseArgs := []string{
		"run",
		"-v", fmt.Sprintf("%s:%s", filepath.Join(dir, "workdir"), "/workdir"),
		"-e", fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", os.Getenv("AWS_ACCESS_KEY_ID")),
		"-e", fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", os.Getenv("AWS_SECRET_ACCESS_KEY")),
		"-e", "KUBECONFIG=/workdir/.shipyard/config",
		"--entrypoint", entrypoint,
	}
	baseArgs = append(baseArgs, args...)

	cmd := exec.Command("docker", baseArgs...)
	cmd.Stdout = out
	cmd.Stderr = out

	return cmd.Run()
}
