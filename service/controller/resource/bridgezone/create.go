package bridgezone

import (
	"context"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	baseDomain := key.ClusterBaseDomain(cr)
	intermediateZone := "k8s." + baseDomain
	finalZone := key.ClusterID(&cr) + ".k8s." + baseDomain

	guest, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	g := &errgroup.Group{}

	var intermediateZoneID HostedZoneID
	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting intermediate zone IDs")

		hostedZoneID, err := r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zones not found")

			return microerror.Mask(err)
		} else if err != nil {
			return microerror.Mask(err)
		}
		intermediateZoneID = hostedZoneID

		r.logger.LogCtx(ctx, "level", "debug", "message", "got intermediate zone IDs")

		return nil
	})

	var finalZoneID HostedZoneID
	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone IDs")

		hostedZoneID, err := r.findHostedZoneID(ctx, guest, finalZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "final zones not found")

			return microerror.Mask(err)
		} else if err != nil {
			return microerror.Mask(err)
		}
		finalZoneID = hostedZoneID

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone IDs")

		return nil
	})

	err = g.Wait()
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var finalPublicZoneRecords []*route53.ResourceRecord
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final public zone name servers")

		nameServers, _, err := r.getNameServersAndTTL(ctx, guest, finalZoneID.Public, finalZone)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, ns := range nameServers {
			copy := ns
			v := &route53.ResourceRecord{
				Value: &copy,
			}
			finalPublicZoneRecords = append(finalPublicZoneRecords, v)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final public zone name servers")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring final public zone delegation from intermediate zone")

		upsert := route53.ChangeActionUpsert
		ns := route53.RRTypeNs
		ttl := int64(300)

		in := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: &upsert,
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name:            &finalZone,
							Type:            &ns,
							TTL:             &ttl,
							ResourceRecords: finalPublicZoneRecords,
						},
					},
				},
			},
			HostedZoneId: &intermediateZoneID.Public,
		}
		_, err := defaultGuest.ChangeResourceRecordSetsWithContext(ctx, in)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured final public zone delegation from intermediate zone")
	}

	// if there is private internmediate zone - handle private records for it as well
	if intermediateZoneID.Private != "" {
		var finalPrivateZoneRecords []*route53.ResourceRecord
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "getting final private zone name servers")

			nameServers, _, err := r.getNameServersAndTTL(ctx, guest, finalZoneID.Private, finalZone)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, ns := range nameServers {
				copy := ns
				v := &route53.ResourceRecord{
					Value: &copy,
				}
				finalPrivateZoneRecords = append(finalPrivateZoneRecords, v)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "got final private zone name servers")
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring final private zone delegation from intermediate zone")

			upsert := route53.ChangeActionUpsert
			ns := route53.RRTypeNs
			ttl := int64(300)

			in := &route53.ChangeResourceRecordSetsInput{
				ChangeBatch: &route53.ChangeBatch{
					Changes: []*route53.Change{
						{
							Action: &upsert,
							ResourceRecordSet: &route53.ResourceRecordSet{
								Name:            &finalZone,
								Type:            &ns,
								TTL:             &ttl,
								ResourceRecords: finalPrivateZoneRecords,
							},
						},
					},
				},
				HostedZoneId: &intermediateZoneID.Private,
			}
			_, err := defaultGuest.ChangeResourceRecordSetsWithContext(ctx, in)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "ensured final private zone delegation from intermediate zone")
		}
	}

	return nil
}
