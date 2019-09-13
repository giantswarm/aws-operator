// Package tcnpazs implements a resource to gather all private subnets for the
// configured availability zones of a node pool. Like the clusterazs resource,
// we need logic to take the node pool subnet allocated by the ipam resource and
// split it according to the configured availability zones. We then have 1, 2 or
// 4 private subnet CIDRs we put into the controller context for further use in
// the tcnp resource. Note that the availability zones of a node pool cannot be
// updated upon creation due to the network splitting. In order to change
// availability zones one must delete and create node pools accordingly.
package tcnpazs

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "tcnpazsv29patch1"
)

type Config struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

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
