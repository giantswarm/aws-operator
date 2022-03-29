package tccpnatgateways

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpnatgateways"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
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
