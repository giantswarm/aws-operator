package tcnp

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/service/internal/images"
)

const (
	// Name is the identifier of the resource.
	Name = "tcnp"
)

type Config struct {
	CloudTags cloudtags.Interface
	Detection *changedetection.TCNP
	Images    images.Interface
	Logger    micrologger.Logger

	InstallationName string
}

// Resource implements the TCNP resource, which stands for Tenant Cluster Data
// Plane. We manage a dedicated Cloud Formation stack for each node pool.
type Resource struct {
	cloudtags cloudtags.Interface
	detection *changedetection.TCNP
	images    images.Interface
	logger    micrologger.Logger

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.CloudTags == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudTags must not be empty", config)
	}
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Images == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Images must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		cloudtags: config.CloudTags,
		detection: config.Detection,
		images:    config.Images,
		logger:    config.Logger,

		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
