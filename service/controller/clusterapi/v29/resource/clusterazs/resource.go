// Package clusterazs implements a resource to gather all distinct availability
// zones for a tenant cluster.
package clusterazs

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "clusterazsv29"
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
