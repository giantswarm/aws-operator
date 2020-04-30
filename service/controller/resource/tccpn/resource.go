package tccpn

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpn"
)

type Config struct {
	G8sClient versioned.Interface
	Detection *changedetection.TCCPN
	HAMaster           hamaster.Interface
	Logger    micrologger.Logger

	InstallationName string
	Route53Enabled   bool
}

// Resource implements the TCCPN resource, which stands for Tenant Cluster
// Control Plane Node. We manage a dedicated Cloud Formation stack for each node
// pool.
type Resource struct {
	g8sClient versioned.Interface
	detection *changedetection.TCCPN
	haMaster           hamaster.Interface
	logger    micrologger.Logger

	installationName string
	route53Enabled   bool
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.HAMaster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HAMaster must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		detection: config.Detection,
		haMaster: config.HAMaster,
		logger:    config.Logger,

		installationName: config.InstallationName,
		route53Enabled:   config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
