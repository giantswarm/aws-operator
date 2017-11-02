package cluster

import (
	"os"

	"github.com/giantswarm/e2e-harness/pkg/runner"
	"github.com/giantswarm/micrologger"
)

type Cluster struct {
	logger        micrologger.Logger
	runner        runner.Runner
	remoteCluster bool
}

func New(logger micrologger.Logger, runner runner.Runner, remoteCluster bool) *Cluster {
	return &Cluster{
		logger:        logger,
		runner:        runner,
		remoteCluster: remoteCluster,
	}
}

// Create is a Task that creates a remote cluster.
func (c *Cluster) Create() error {
	return c.clusterAction("shipyard -action=start")
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

	return err
}
