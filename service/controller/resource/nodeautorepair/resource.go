package nodeautorepair

import (
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "nodeautorepair"
)

type Config struct {
	Logger micrologger.Logger

	Enabled           bool
	NotReadyThreshold time.Duration
}

type Resource struct {
	logger micrologger.Logger

	enabled           bool
	notReadyThreshold time.Duration
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.NotReadyThreshold == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.NotReadyThreshold must not be zero", config)
	}

	r := &Resource{
		logger:            config.Logger,
		enabled:           config.Enabled,
		notReadyThreshold: config.NotReadyThreshold,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
