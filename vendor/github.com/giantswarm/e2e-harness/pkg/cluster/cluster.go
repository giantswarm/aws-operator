package cluster

import (
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/runner"
)

type Cluster struct {
	logger        micrologger.Logger
	runner        runner.Runner
	fs            afero.Fs
	remoteCluster bool
}

func New(logger micrologger.Logger, fs afero.Fs, runner runner.Runner, remoteCluster bool) *Cluster {
	return &Cluster{
		logger:        logger,
		runner:        runner,
		fs:            fs,
		remoteCluster: remoteCluster,
	}
}

// Create is a Task that creates a remote cluster or, if we
// are using a local one, puts in place the required files for
// later access to it
func (c *Cluster) Create() error {
	if c.remoteCluster {
		err := c.clusterAction("shipyard -action=start")
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	usr, err := user.Current()
	if err != nil {
		return microerror.Mask(err)
	}

	err = c.copyMinikubeAssets(usr.HomeDir)
	if err != nil {
		return microerror.Mask(err)
	}
	err = c.setupMinikubeConfig(usr.HomeDir)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

// Delete is a Task that gets rid of a remote cluster.
func (c *Cluster) Delete() error {
	return c.clusterAction("shipyard -action=stop")
}

func (c *Cluster) clusterAction(command string) error {
	if !c.remoteCluster {
		return nil
	}
	err := c.runner.Run(os.Stdout, command)

	return microerror.Mask(err)
}

// copyMinikubeAssets copies all the files found in $HOME/.minikube to
// the e2e-harness workdir (so that they will be accessible from the test
// container)
func (c *Cluster) copyMinikubeAssets(homeDir string) error {
	c.logger.Log("info", "Making minikube assets accessible for the test container")

	originDir := filepath.Join(homeDir, ".minikube")
	baseDir, err := harness.BaseDir()
	if err != nil {
		return microerror.Mask(err)
	}
	targetDir := filepath.Join(baseDir, "workdir", ".minikube")

	// copy minikube directory
	walkFn := func(path string, info os.FileInfo, err error) error {
		targetPath := strings.Replace(path, originDir, targetDir, 1)
		if info.IsDir() {
			return c.fs.MkdirAll(targetPath, os.ModePerm)
		}
		return c.copyFile(path, targetPath)
	}
	err = filepath.Walk(originDir, walkFn)
	if err != nil {
		return microerror.Mask(err)
	}

	// copy kube config (assumes the current context is minukube)
	origKubeCfg := filepath.Join(homeDir, ".kube", "config")
	targetKubeCfg, err := getMinikubeConfigPath()
	if err != nil {
		return microerror.Mask(err)
	}
	targetKubeCfgDir := filepath.Dir(targetKubeCfg)
	if err := c.fs.MkdirAll(targetKubeCfgDir, os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := c.copyFile(origKubeCfg, targetKubeCfg); err != nil {
		return microerror.Mask(err)
	}
	return nil
}

// setupMinikubeConfig replaces $HOME/.minukube in the k8s config
// file (as seen by the container where all the commands are going to
// be executed) by the path where the certificates can be found (again,
// from the container point of view).
func (c *Cluster) setupMinikubeConfig(homeDir string) error {
	c.logger.Log("info", "Setting up minikube config for the test container")

	// the default k8s config file references the required certificates
	// to access minikube using $HOME/.minikube, we store this in originDir
	originDir := filepath.Join(homeDir, ".minikube")

	// path is the actual location of the k8s config file that will be used from the
	// test container
	path, err := getMinikubeConfigPath()
	if err != nil {
		return microerror.Mask(err)
	}
	afs := &afero.Afero{Fs: c.fs}

	// circumvent umask settings, by assigning the right permissions to shipyard dir
	// (afero creates a .lock file on read..)
	baseDir, err := harness.BaseDir()
	if err != nil {
		return microerror.Mask(err)
	}
	shipyardDir := filepath.Join(baseDir, "workdir", ".shipyard")
	err = c.fs.Chmod(shipyardDir, 0777)
	if err != nil {
		return microerror.Mask(err)
	}

	read, err := afs.ReadFile(path)
	if err != nil {
		return microerror.Mask(err)
	}

	// targetDir has the path where minikube certificates are stored as seen from
	// the test container
	targetDir := filepath.Join("/workdir", ".minikube")

	newContents := strings.Replace(string(read), originDir, targetDir, -1)

	err = afs.WriteFile(path, []byte(newContents), 0)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

// getMinikubeConfigPath returns the actual path of the k8s config file that
// will be used by the test container (path from the point of view of the
// executing e2e-harness binary, not the test container).
func getMinikubeConfigPath() (string, error) {
	baseDir, err := harness.BaseDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(baseDir, "workdir", ".shipyard", "config")

	return path, nil
}

func (c *Cluster) copyFile(orig, dst string) error {
	in, err := c.fs.Open(orig)
	if err != nil {
		return microerror.Mask(err)
	}
	defer in.Close()
	out, err := c.fs.Create(dst)
	if err != nil {
		return microerror.Mask(err)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return microerror.Mask(err)
	}
	err = out.Sync()

	return microerror.Mask(err)
}
