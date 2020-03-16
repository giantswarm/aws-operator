package tccpvpcid

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpvpcid"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
