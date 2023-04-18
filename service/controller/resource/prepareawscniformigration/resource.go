package prepareawscniformigration

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "prepareawscniformigration"
)

type Config struct {
	CtrlClient     client.Client
	Logger         micrologger.Logger
	RegistryDomain string
}

// Resource that ensures the `aws-node` daemonset is configured correctly for migration to cilum
type Resource struct {
	ctrlClient     client.Client
	logger         micrologger.Logger
	registryDomain string
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	r := &Resource{
		ctrlClient:     config.CtrlClient,
		logger:         config.Logger,
		registryDomain: config.RegistryDomain,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
