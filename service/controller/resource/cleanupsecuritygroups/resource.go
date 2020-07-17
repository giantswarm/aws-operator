package cleanupsecuritygroups

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/recorder"
)

const (
	Name = "cleanupsecuritygroups"
)

type Config struct {
	Event  recorder.Interface
	Logger micrologger.Logger
}

type Resource struct {
	event  recorder.Interface
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		event:  config.Event,
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
