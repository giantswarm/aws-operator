package cphostedzone

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/route53"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Config struct {
	Logger micrologger.Logger

	Route53Enabled bool
}

type HostedZone struct {
	logger micrologger.Logger

	cachedCPHostedZoneID         string
	cachedCPInternalHostedZoneID string

	mutex sync.Mutex

	route53Enabled bool
}

func New(config Config) (*HostedZone, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	h := &HostedZone{
		logger: config.Logger,

		cachedCPHostedZoneID: "",
		mutex:                sync.Mutex{},

		route53Enabled: config.Route53Enabled,
	}

	return h, nil
}

func (h *HostedZone) Search(ctx context.Context, obj interface{}) (string, string, error) {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	if !h.route53Enabled {
		h.logger.Debugf(ctx, "route53 disabled")
		h.logger.Debugf(ctx, "canceling resource")
		return "", "", nil
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	cpHostedZoneID, cpInternalHostedZoneID, err := h.lookup(ctx, cc.Client.ControlPlane.AWS.Route53, cr)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	return cpHostedZoneID, cpInternalHostedZoneID, nil
}

func (h *HostedZone) lookup(ctx context.Context, client Route53, cr infrastructurev1alpha3.AWSCluster) (cpHostedZoneID, cpInternalHostedZoneID string, err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// We check if we have CP public HostedZone info cached.
	{
		h.logger.Debugf(ctx, "finding cached CP public HostedZone ID")

		if h.cachedCPHostedZoneID != "" {
			h.logger.Debugf(ctx, "found cached CP public HostedZone ID %#q", h.cachedCPHostedZoneID)
			cpHostedZoneID = h.cachedCPHostedZoneID
		}

		h.logger.Debugf(ctx, "did not find cached CP public HostedZone ID")
	}

	// We check if we have CP public HostedZone info cached.
	{
		h.logger.Debugf(ctx, "finding cached CP private HostedZone ID")

		if h.cachedCPInternalHostedZoneID != "" {
			h.logger.Debugf(ctx, "found cached CP private HostedZone ID %#q", h.cachedCPInternalHostedZoneID)
			cpInternalHostedZoneID = h.cachedCPInternalHostedZoneID
		}

		h.logger.Debugf(ctx, "did not find cached CP private HostedZone ID")
	}

	if cpHostedZoneID != "" && cpInternalHostedZoneID != "" {
		return cpHostedZoneID, cpInternalHostedZoneID, nil
	}

	// We do not have a cached CP HostedZones Info for the requested
	// installation. So we look it up.
	h.logger.Debugf(ctx, "finding CP HostedZone IDs")

	hostedZonesInput := &route53.ListHostedZonesByNameInput{}

	o, err := client.ListHostedZonesByName(hostedZonesInput)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	baseDomain := fmt.Sprintf("%s.", key.ClusterBaseDomain(cr))

	for _, zone := range o.HostedZones {
		if *zone.Name == baseDomain && !*zone.Config.PrivateZone {
			h.logger.Debugf(ctx, "found CP public HostedZone ID %#q", cpHostedZoneID)
			cpHostedZoneID = *zone.Id

			h.logger.Debugf(ctx, "caching CP public HostedZone ID")
			h.cachedCPHostedZoneID = cpHostedZoneID
			h.logger.Debugf(ctx, "cached CP public HostedZone ID")

		}

		if *zone.Name == baseDomain && *zone.Config.PrivateZone {
			h.logger.Debugf(ctx, "found CP private HostedZone ID %#q", cpInternalHostedZoneID)
			cpInternalHostedZoneID = *zone.Id

			h.logger.Debugf(ctx, "caching CP private HostedZone ID")
			h.cachedCPInternalHostedZoneID = cpInternalHostedZoneID
			h.logger.Debugf(ctx, "cached CP private HostedZone ID")
		}
	}

	// Fail only if public hosted zone is missing as having private hosted zone in CP is not a requirement
	if cpHostedZoneID == "" {
		return "", "", microerror.Maskf(executionFailedError, "failed to find CP public HostedZone ID for base domain %#q", baseDomain)
	}

	return cpHostedZoneID, cpInternalHostedZoneID, nil
}
