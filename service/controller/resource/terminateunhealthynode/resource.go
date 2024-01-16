package terminateunhealthynode

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	event "github.com/giantswarm/aws-operator/v16/service/internal/recorder"
)

const (
	Name = "terminateunhealthynode"
)

type Config struct {
	Event  event.Interface
	Logger micrologger.Logger
}

type Resource struct {
	event  event.Interface
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
