package tccpf

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v14/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v14/service/internal/cphostedzone"
	"github.com/giantswarm/aws-operator/v14/service/internal/recorder"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpf"
)

type Config struct {
	Detection  *changedetection.TCCPF
	Event      recorder.Interface
	HostedZone *cphostedzone.HostedZone
	Logger     micrologger.Logger

	InstallationName string
	Route53Enabled   bool
}

// Resource implements the TCCPF resource, which stands for Tenant Cluster
// Control Plane Finalizer. This was formerly known as the host main stack. We
// manage a dedicated CF stack for the record sets and routing tables setup.
type Resource struct {
	detection  *changedetection.TCCPF
	event      recorder.Interface
	hostedZone *cphostedzone.HostedZone
	logger     micrologger.Logger

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
	if config.HostedZone == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostedZone must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		detection:  config.Detection,
		event:      config.Event,
		hostedZone: config.HostedZone,
		logger:     config.Logger,

		installationName: config.InstallationName,
		route53Enabled:   config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
