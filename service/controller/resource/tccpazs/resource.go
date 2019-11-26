// Package tccpazs implements a resource to gather all distinct availability
// zones for a tenant cluster.
package tccpazs

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "tccpazs"
)

type Config struct {
	CMAClient     clientset.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (infrastructurev1alpha2.Cluster, error)
}

type Resource struct {
	cmaClient     clientset.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (infrastructurev1alpha2.Cluster, error)
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
