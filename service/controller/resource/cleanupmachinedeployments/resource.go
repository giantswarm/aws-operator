package cleanupmachinedeployments

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	event "github.com/giantswarm/aws-operator/v12/service/internal/recorder"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "cleanupmachinedeployments"
)

type Config struct {
	Event      event.Interface
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger
}

type Resource struct {
	event      event.Interface
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		event:      config.Event,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
