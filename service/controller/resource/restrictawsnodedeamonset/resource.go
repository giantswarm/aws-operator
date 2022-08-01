package restrictawsnodedaemonset

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "restrictawsnodedaemonset"
)

type Config struct {
	Logger micrologger.Logger
}

// Resource that ensures the `aws-node` daemonset in the WC only creates pods in old nodes, not upgraded ones.
type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
