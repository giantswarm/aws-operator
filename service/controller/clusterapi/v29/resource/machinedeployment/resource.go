// Package machinedeployment implements a temporary resource to fetch the single
// MachineDeployment of a tenant cluster. Once we separate the tenant cluster's
// data plane in the cloud formation structure, this resource here will be
// deleted again, because the MachineDeployment will be reconciled with its own
// controller and resources.
package machinedeployment

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "machinedeploymentv29"
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
