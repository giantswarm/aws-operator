package tccpn

import (
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/images"
	event "github.com/giantswarm/aws-operator/service/internal/recorder"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpn"
)

type Config struct {
	Detection *changedetection.TCCPN
	Event     event.Interface
	HAMaster  hamaster.Interface
	Images    images.Interface
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	InstallationName string
	Route53Enabled   bool
}

// Resource implements the TCCPN resource, which stands for Tenant Cluster
// Control Plane Node. We manage a dedicated Cloud Formation stack for each node
// pool.
type Resource struct {
	detection *changedetection.TCCPN
	event     event.Interface
	haMaster  hamaster.Interface
	images    images.Interface
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	installationName string
	route53Enabled   bool
}

func New(config Config) (*Resource, error) {
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.HAMaster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HAMaster must not be empty", config)
	}
	if config.Images == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Images must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		detection: config.Detection,
		event:     config.Event,
		haMaster:  config.HAMaster,
		images:    config.Images,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		installationName: config.InstallationName,
		route53Enabled:   config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
