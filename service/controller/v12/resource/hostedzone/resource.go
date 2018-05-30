package hostedzone

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	controllercontext "github.com/giantswarm/aws-operator/service/controller/v12/context"
	"github.com/giantswarm/aws-operator/service/controller/v12/key"
)

const (
	name = "hostedzonev12"
)

type Config struct {
	HostRoute53 *route53.Route53
	Logger      micrologger.Logger
}

type Resource struct {
	hostRoute53 *route53.Route53
	logger      micrologger.Logger
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

		apiDomain     = strings.TrimSuffix(key.HostedZoneNameAPI(customObject), ".")
		etcdDomain    = strings.TrimSuffix(key.HostedZoneNameEtcd(customObject), ".")
		ingressDomain = strings.TrimSuffix(key.HostedZoneNameIngress(customObject), ".")
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

			switch hzName {
			case apiDomain:
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found hosted zone ID %q for domain %q", hzName, hzID))
				controllerCtx.Status.HostedZones.API.ID = hzID
				apiFound = true
			case etcdDomain:
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found hosted zone ID %q for domain %q", hzName, hzID))
				controllerCtx.Status.HostedZones.Etcd.ID = hzID
				etcdFound = true
			case ingressDomain:
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found hosted zone ID %q for domain %q", hzName, hzID))
				controllerCtx.Status.HostedZones.Ingress.ID = hzID
				ingressFound = true
			default:
				// Hosted Zone doesn't match any of desired
				// domains. Continue.
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
		return microerror.Maskf(hostedZoneNotFoundError, "domain = %q", apiDomain)
	}
	if !etcdFound {
		return microerror.Maskf(hostedZoneNotFoundError, "domain = %q", etcdDomain)
	}
	if !ingressFound {
		return microerror.Maskf(hostedZoneNotFoundError, "domain = %q", ingressDomain)
	}

	return nil
}
