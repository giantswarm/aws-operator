package tcnp

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
)

const (
	// Name is the identifier of the resource.
	Name = "tcnp"
)

type Config struct {
	G8sClient versioned.Interface
	Detection *changedetection.TCNP
	Logger    micrologger.Logger

	EncrypterBackend string
	InstallationName string
}

// Resource implements the TCNP resource, which stands for Tenant Cluster Data
// Plane. We manage a dedicated Cloud Formation stack for each node pool.
type Resource struct {
	g8sClient versioned.Interface
	detection *changedetection.TCNP
	logger    micrologger.Logger

	encrypterBackend string
	installationName string
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		detection: config.Detection,
		logger:    config.Logger,

		encrypterBackend: config.EncrypterBackend,
		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
