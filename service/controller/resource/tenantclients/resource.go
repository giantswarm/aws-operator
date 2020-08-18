package tenantclients

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v3/pkg/tenantcluster"
)

const (
	Name = "tenantclients"
)

type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface
}

type Resource struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
		tenant: config.Tenant,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
