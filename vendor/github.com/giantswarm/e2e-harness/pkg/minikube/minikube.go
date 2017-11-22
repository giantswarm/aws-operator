package minikube

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2e-harness/pkg/builder"
	"github.com/giantswarm/e2e-harness/pkg/harness"
)

type Minikube struct {
	logger   micrologger.Logger
	builder  builder.Builder
	imageTag string
}

func New(logger micrologger.Logger, builder builder.Builder, tag string) *Minikube {
	return &Minikube{
		logger:   logger,
		builder:  builder,
		imageTag: tag,
	}
}

// BuildImages is a Task that build the required images for both the main
// project and the e2e containers using the minikube docker environment.
func (m *Minikube) BuildImages() error {
	m.logger.Log("info", "Getting minikube docker environment")
	env, err := m.getDockerEnv()
	dir, err := os.Getwd()
	if err != nil {
		return microerror.Mask(err)
	}

	name := harness.GetProjectName()

	image := fmt.Sprintf("quay.io/giantswarm/%s", name)
	m.logger.Log("info", "Building image "+image)
	if err := m.buildImage(name, dir, image, env); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (m *Minikube) getDockerEnv() ([]string, error) {
	cmd := exec.Command("minikube", "docker-env")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return []string{}, microerror.Mask(err)
	}
	if err := cmd.Start(); err != nil {
		return []string{}, microerror.Mask(err)
	}

	var env []string

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "export") {
			parts := strings.Fields(scanner.Text())
			entry := strings.Replace(parts[1], `"`, "", -1)
			env = append(env, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return env, microerror.Mask(err)
	}

	if err := cmd.Wait(); err != nil {
		return []string{}, microerror.Mask(err)
	}

	return env, nil
}

func (m *Minikube) buildImage(binaryName, path, imageName string, env []string) error {
	if err := m.builder.Build(ioutil.Discard, imageName, path, m.imageTag, env); err != nil {
		fmt.Println("error building image", imageName)
		return microerror.Mask(err)
	}
	return nil
}
