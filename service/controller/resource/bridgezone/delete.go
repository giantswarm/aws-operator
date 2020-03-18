package bridgezone

import (
	"context"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
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

	_, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID HostedZoneID
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting intermediate zone ID")

		intermediateZoneID, err = r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got intermediate zone ID")
	}

	var finalPublicZoneTTL int64
	var finalPublicZoneRecords []*route53.ResourceRecord
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final public zone delegation name servers and TTL from intermediate zone")

		nameServers, ttl, err := r.getNameServersAndTTL(ctx, defaultGuest, intermediateZoneID.Public, finalZone)
		if IsNotFound(err) {
			// Delegation may be already deleted. It must be handled.
			r.logger.LogCtx(ctx, "level", "debug", "message", "final public zone delegation not found in intermediate zone")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		finalPublicZoneTTL = ttl

		for _, ns := range nameServers {
			copy := ns
			v := &route53.ResourceRecord{
				Value: &copy,
			}
			finalPublicZoneRecords = append(finalPublicZoneRecords, v)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final public zone delegation name servers and TTL from intermediate zone")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of final public zone delegation from intermediate zone")

		delete := route53.ChangeActionDelete
		ns := route53.RRTypeNs

		in := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: &delete,
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name:            &finalZone,
							Type:            &ns,
							TTL:             &finalPublicZoneTTL,
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

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured deletion of final public zone delegation from intermediate zone")
	}

	// if there is private internmediate zone - handle private records for it as well
	if intermediateZoneID.Private != "" {

		var finalPrivateZoneTTL int64
		var finalPrivateZoneRecords []*route53.ResourceRecord
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "getting final private zone delegation name servers and TTL from intermediate zone")

			nameServers, ttl, err := r.getNameServersAndTTL(ctx, defaultGuest, intermediateZoneID.Private, finalZone)
			if IsNotFound(err) {
				// Delegation may be already deleted. It must be handled.
				r.logger.LogCtx(ctx, "level", "debug", "message", "final private zone delegation not found in intermediate zone")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}

			finalPrivateZoneTTL = ttl

			for _, ns := range nameServers {
				copy := ns
				v := &route53.ResourceRecord{
					Value: &copy,
				}
				finalPrivateZoneRecords = append(finalPrivateZoneRecords, v)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "got final private zone delegation name servers and TTL from intermediate zone")
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of final private zone delegation from intermediate zone")

			delete := route53.ChangeActionDelete
			ns := route53.RRTypeNs

			in := &route53.ChangeResourceRecordSetsInput{
				ChangeBatch: &route53.ChangeBatch{
					Changes: []*route53.Change{
						{
							Action: &delete,
							ResourceRecordSet: &route53.ResourceRecordSet{
								Name:            &finalZone,
								Type:            &ns,
								TTL:             &finalPrivateZoneTTL,
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

			r.logger.LogCtx(ctx, "level", "debug", "message", "ensured deletion of final public zone delegation from intermediate zone")
		}
	}

	return nil
}
