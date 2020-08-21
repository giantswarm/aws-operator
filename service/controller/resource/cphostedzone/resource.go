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

	cachedCPHostedZoneID string
	mutex                sync.Mutex

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
		hostedZoneID, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.Route53, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.ControlPlane.HostedZone.ID = hostedZoneID
	}

	return nil
}

func (r *Resource) lookup(ctx context.Context, client Route53, cr infrastructurev1alpha2.AWSCluster) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// We check if we have CP HostedZone info cached.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cached CP public HostedZone ID")

		if r.cachedCPHostedZoneID != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found cached CP public HostedZone ID %#q", r.cachedCPHostedZoneID)
			return r.cachedCPHostedZoneID, nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cached CP public HostedZone ID")
	}

	// We do not have a cached CP public HostedZone Info for the requested
	// installation. So we look it up.
	var cpHostedZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding CP public HostedZone ID")

		hostedZonesInput := &route53.ListHostedZonesByNameInput{}

		o, err := client.ListHostedZonesByName(hostedZonesInput)
		if err != nil {
			return "", microerror.Mask(err)
		}

		baseDomain := fmt.Sprintf("%s.", key.ClusterBaseDomain(cr))

		for _, zone := range o.HostedZones {
			if *zone.Name == baseDomain && !*zone.Config.PrivateZone {
				cpHostedZoneID = *zone.Id
				break
			}
		}

		if cpHostedZoneID == "" {
			return "", microerror.Maskf(executionFailedError, "failed to find CP public HostedZone ID for base domain %#q", baseDomain)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found CP public HostedZone ID %#q", cpHostedZoneID))
	}

	// At this point we found a public HostedZone ID info and we cache it.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "caching CP public HostedZone ID")
		r.cachedCPHostedZoneID = cpHostedZoneID
		r.logger.LogCtx(ctx, "level", "debug", "message", "cached CP public HostedZone ID")
	}

	return cpHostedZoneID, nil
}
