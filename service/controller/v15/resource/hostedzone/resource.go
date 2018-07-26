package hostedzone

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v15/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

const (
	name = "hostedzonev15"
)

type Config struct {
	HostRoute53 *route53.Route53
	Logger      micrologger.Logger

	Route53Enabled bool
}

type Resource struct {
	hostRoute53 *route53.Route53
	logger      micrologger.Logger

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if config.HostRoute53 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostRoute53 must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		hostRoute53: config.HostRoute53,
		logger:      config.Logger,

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.setStatus(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	err := r.setStatus(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// setStatus searches for HostedZone in AWS API by name for their IDs. Those
// IDs are set in controller context status for further use.
func (r *Resource) setStatus(ctx context.Context, obj interface{}) error {
	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled, skipping execution")
		return nil
	}

	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	controllerCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for host cluster hosted zone IDs in AWS API")

	var (
		apiFound     = false
		etcdFound    = false
		ingressFound = false

		apiZone     = strings.TrimSuffix(key.HostedZoneNameAPI(customObject), ".")
		etcdZone    = strings.TrimSuffix(key.HostedZoneNameEtcd(customObject), ".")
		ingressZone = strings.TrimSuffix(key.HostedZoneNameIngress(customObject), ".")
	)

	var marker *string
	for {
		in := &route53.ListHostedZonesInput{
			Marker: marker,
		}

		out, err := r.hostRoute53.ListHostedZones(in)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, hz := range out.HostedZones {
			if hz.Name == nil || hz.Id == nil {
				continue
			}

			hzName := *hz.Name
			hzName = strings.TrimSuffix(hzName, ".")
			hzID := *hz.Id

			if hzName == apiZone {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found api hosted zone ID %q for domain %q", hzID, hzName))
				controllerCtx.Status.HostedZones.API.ID = hzID
				apiFound = true
			}
			if hzName == etcdZone {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found etcd hosted zone ID %q for domain %q", hzID, hzName))
				controllerCtx.Status.HostedZones.Etcd.ID = hzID
				etcdFound = true
			}
			if hzName == ingressZone {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found ingress hosted zone ID %q for domain %q", hzID, hzName))
				controllerCtx.Status.HostedZones.Ingress.ID = hzID
				ingressFound = true
			}
		}

		// If all IDs are found stop here.
		allFound := apiFound && etcdFound && ingressFound
		if allFound {
			break
		}

		// If not all IDs are found, try to search next page.
		if out.IsTruncated == nil || !*out.IsTruncated {
			break
		}
		marker = out.Marker
	}

	if !apiFound {
		return microerror.Maskf(hostedZoneNotFoundError, "zone = %q", apiZone)
	}
	if !etcdFound {
		return microerror.Maskf(hostedZoneNotFoundError, "zone = %q", etcdZone)
	}
	if !ingressFound {
		return microerror.Maskf(hostedZoneNotFoundError, "zone = %q", ingressZone)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found all host cluster hosted zone IDs")
	return nil
}
