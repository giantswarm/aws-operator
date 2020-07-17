package secretfinalizer

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/service/internal/recorder"
)

const (
	Name = "secretfinalizer"
)

const (
	secretFinalizer = "aws-operator.giantswarm.io/secretfinalizer"
)

type Config struct {
	Event     recorder.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	event     recorder.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		event:     config.Event,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r Resource) Name() string {
	return Name
}
