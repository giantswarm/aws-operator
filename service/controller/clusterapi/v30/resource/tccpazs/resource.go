// Package tccpazs implements a resource to gather all distinct availability
// zones for a tenant cluster.
package tccpazs

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "tccpazsv30"
)

type Config struct {
	CMAClient     clientset.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

type Resource struct {
	cmaClient     clientset.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		cmaClient:     config.CMAClient,
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
