package cleanuprecordsets

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/recorder"
)

const (
	Name = "cleanuprecordsetsv31"
)

type Config struct {
	Event  recorder.Interface
	Logger micrologger.Logger

	Route53Enabled bool
}

type Resource struct {
	event  recorder.Interface
	logger micrologger.Logger

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
