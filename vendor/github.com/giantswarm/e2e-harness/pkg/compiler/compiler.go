package compiler

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2e-harness/pkg/harness"
)

type Compiler struct {
	logger micrologger.Logger
}

func New(logger micrologger.Logger) *Compiler {
	return &Compiler{
		logger: logger,
	}
}

// CompileMain is a Task that builds the main binary.
func (c *Compiler) CompileMain() error {
	dir, err := os.Getwd()
	if err != nil {
		return microerror.Mask(err)
	}

	name := harness.GetProjectName()

	c.logger.Log("info", "Compiling binary "+name)
	if err := c.compileMain(name, dir); err != nil {
		c.logger.Log("info", "error compiling binary "+name)
		return microerror.Mask(err)
	}

	return nil
}

// CompileTests is a Task that builds the tests binary.
func (c *Compiler) CompileTests() error {
	dir, err := os.Getwd()
	if err != nil {
		return microerror.Mask(err)
	}

	name := harness.GetProjectName()

	e2eBinary := name + "-e2e"
	e2eDir := filepath.Join(dir, "integration")
	c.logger.Log("info", "Compiling binary "+e2eBinary)
	if err := c.compileTests(e2eBinary, e2eDir); err != nil {
		c.logger.Log("info", "error compiling binary "+e2eBinary)
		return microerror.Mask(err)
	}
	return nil
}

func (c *Compiler) compileMain(binaryName, path string) error {
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	cmd.Dir = path

	return cmd.Run()
}

func (c *Compiler) compileTests(binaryName, path string) error {
	cmd := exec.Command("go", "test", "-c", "-o", binaryName, "-tags", "k8srequired", ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	cmd.Dir = path

	return cmd.Run()
}
