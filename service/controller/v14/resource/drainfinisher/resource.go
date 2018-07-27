package drainfinisher

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "drainfinisherv14"
)

type ResourceConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newResource := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
