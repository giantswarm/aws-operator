package cleanuprecordsets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var hostedZones []*route53.HostedZone
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding hosted zones")

		hostedZonesInput := &route53.ListHostedZonesByNameInput{}

		o, err := cc.Client.TenantCluster.AWS.Route53.ListHostedZonesByName(hostedZonesInput)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, zone := range o.HostedZones {
			baseDomain := fmt.Sprintf("%s.", key.TenantClusterBaseDomain(cr))

			if *zone.Name == baseDomain {
				hostedZones = append(hostedZones, zone)

				r.logger.LogCtx(ctx, "level", "debug", "message", "found hosted zones")
			}
		}
	}

	for _, hostedZone := range hostedZones {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding record sets in hosted zone %#q", *hostedZone.Id))

		recordSetsInput := &route53.ListResourceRecordSetsInput{
			HostedZoneId: hostedZone.Id,
		}

		o, err := cc.Client.TenantCluster.AWS.Route53.ListResourceRecordSets(recordSetsInput)
		if err != nil {
			return microerror.Mask(err)
		}

		resourceRecordSets := o.ResourceRecordSets

		managedRecordSets := key.ManagedRecordSets(cr)
		route53Changes := []*route53.Change{}
		for _, rr := range resourceRecordSets {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("looking for non-managed record sets in hosted zone %#q", *hostedZone.Id))

			if !stringInSlice(*rr.Name, managedRecordSets) {
				route53Change := &route53.Change{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						AliasTarget:     rr.AliasTarget,
						Name:            rr.Name,
						ResourceRecords: rr.ResourceRecords,
						TTL:             rr.TTL,
						Type:            rr.Type,
						Weight:          rr.Weight,
						SetIdentifier:   rr.SetIdentifier,
					},
				}

				route53Changes = append(route53Changes, route53Change)

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found non-managed record set %#q in hosted zone %#q", *rr.Name, *hostedZone.Id))
			}
		}

		if len(route53Changes) > 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleteting non-managed record sets in hosted zone %#q", *hostedZone.Id))

			changeRecordSetInput := &route53.ChangeResourceRecordSetsInput{
				ChangeBatch: &route53.ChangeBatch{
					Changes: route53Changes,
				},
				HostedZoneId: hostedZone.Id,
			}

			_, err = cc.Client.TenantCluster.AWS.Route53.ChangeResourceRecordSets(changeRecordSetInput)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted non-managed record sets in hosted zone %#q", *hostedZone.Id))
		}
	}

	return nil
}

func stringInSlice(str string, list []string) bool {
	for _, value := range list {
		if value == str {
			return true
		}
	}
	return false
}
