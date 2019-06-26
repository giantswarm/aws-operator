// +build k8srequired

package loadtest

import (
	"context"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type ProviderConfig struct {
	GuestFramework *framework.Guest
	HostFramework  *framework.Host
	Logger         micrologger.Logger

	ClusterID string
}

type Provider struct {
	guestFramework *framework.Guest
	hostFramework  *framework.Host
	logger         micrologger.Logger

	clusterID string
}

func NewProvider(config ProviderConfig) (*Provider, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	p := &Provider{
		guestFramework: config.GuestFramework,
		hostFramework:  config.HostFramework,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return p, nil
}

func (p *Provider) InstallTestApp(ctx context.Context) error {
	return nil
}
