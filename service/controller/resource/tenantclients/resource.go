package tenantclients

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v6/pkg/tenantcluster"
)

const (
	Name = "tenantclients"
)

type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface

	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
}

type Resource struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface

	toClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
		tenant: config.Tenant,

		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
