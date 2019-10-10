package cleanupmachinedeployments

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "cleanupmachinedeploymentsv31"
)

type Config struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

// TODO the whole resource should be moved to cluster-operator as this
// functionality is provider independent and absolutely not AWS specific. As
// soon as we want to implement Node Pools and/or Cluster API for another
// provider we want that to be done for all providers anyway.
//
//     https://github.com/giantswarm/giantswarm/issues/7221
//
type Resource struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
