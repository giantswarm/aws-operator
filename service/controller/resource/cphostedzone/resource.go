package cphostedzone

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/route53"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "cphostedzone"
)

type Config struct {
	Logger micrologger.Logger

	Route53Enabled bool
}

type Resource struct {
	logger micrologger.Logger

	cachedCPHostedZoneID         string
	cachedCPInternalHostedZoneID string

	mutex sync.Mutex

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		cachedCPHostedZoneID: "",
		mutex:                sync.Mutex{},

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addHostedZoneInfoToContext(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		cpHostedZoneID, cpInternalHostedZoneID, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.Route53, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.ControlPlane.HostedZone.ID = cpHostedZoneID
		cc.Status.ControlPlane.InternalHostedZone.ID = cpInternalHostedZoneID
	}

	return nil
}

func (r *Resource) lookup(ctx context.Context, client Route53, cr infrastructurev1alpha2.AWSCluster) (string, string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// We check if we have CP public HostedZone info cached.
	var cpHostedZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cached CP public HostedZone ID")

		if r.cachedCPHostedZoneID != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cached CP public HostedZone ID %#q", r.cachedCPHostedZoneID))
			cpHostedZoneID = r.cachedCPHostedZoneID
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cached CP public HostedZone ID")
	}

	// We check if we have CP public HostedZone info cached.
	var cpInternalHostedZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cached CP private HostedZone ID")

		if r.cachedCPInternalHostedZoneID != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cached CP private HostedZone ID %#q", r.cachedCPInternalHostedZoneID))
			cpInternalHostedZoneID = r.cachedCPInternalHostedZoneID
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cached CP private HostedZone ID")
	}

	if cpHostedZoneID != "" && cpInternalHostedZoneID != "" {
		return cpHostedZoneID, cpInternalHostedZoneID, nil
	}

	// We do not have a cached CP HostedZones Info for the requested
	// installation. So we look it up.
	r.logger.LogCtx(ctx, "level", "debug", "message", "finding CP HostedZone IDs")

	hostedZonesInput := &route53.ListHostedZonesByNameInput{}

	o, err := client.ListHostedZonesByName(hostedZonesInput)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	baseDomain := fmt.Sprintf("%s.", key.ClusterBaseDomain(cr))

	for _, zone := range o.HostedZones {
		if *zone.Name == baseDomain && !*zone.Config.PrivateZone {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found CP public HostedZone ID %#q", cpHostedZoneID))
			cpHostedZoneID = *zone.Id

			r.logger.LogCtx(ctx, "level", "debug", "message", "caching CP public HostedZone ID")
			r.cachedCPHostedZoneID = cpHostedZoneID
			r.logger.LogCtx(ctx, "level", "debug", "message", "cached CP public HostedZone ID")

		}

		if *zone.Name == baseDomain && *zone.Config.PrivateZone {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found CP private HostedZone ID %#q", cpInternalHostedZoneID))
			cpInternalHostedZoneID = *zone.Id

			r.logger.LogCtx(ctx, "level", "debug", "message", "caching CP private HostedZone ID")
			r.cachedCPInternalHostedZoneID = cpInternalHostedZoneID
			r.logger.LogCtx(ctx, "level", "debug", "message", "cached CP private HostedZone ID")
		}
	}

	// Fail only if public hosted zone is missing as having private hosted zone in CP is not a requirement
	if cpHostedZoneID == "" {
		return "", "", microerror.Maskf(executionFailedError, "failed to find CP public HostedZone ID for base domain %#q", baseDomain)
	}

	return cpHostedZoneID, cpInternalHostedZoneID, nil
}
