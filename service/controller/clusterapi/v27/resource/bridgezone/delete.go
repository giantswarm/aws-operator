package bridgezone

import (
	"context"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	finalZone := key.ClusterID(cr) + ".k8s." + baseDomain

	_, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID string
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

	var finalZoneTTL int64
	var finalZoneRecords []*route53.ResourceRecord
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone delegation name servers and TTL from intermediate zone")

		nameServers, ttl, err := r.getNameServersAndTTL(ctx, defaultGuest, intermediateZoneID, finalZone)
		if IsNotFound(err) {
			// Delegation may be already deleted. It must be handled.
			r.logger.LogCtx(ctx, "level", "debug", "message", "final zone delegation not found in intermediate zone")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		finalZoneTTL = ttl

		for _, ns := range nameServers {
			copy := ns
			v := &route53.ResourceRecord{
				Value: &copy,
			}
			finalZoneRecords = append(finalZoneRecords, v)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone delegation name servers and TTL from intermediate zone")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of final zone delegation from intermediate zone")

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
							TTL:             &finalZoneTTL,
							ResourceRecords: finalZoneRecords,
						},
					},
				},
			},
			HostedZoneId: &intermediateZoneID,
		}
		_, err := defaultGuest.ChangeResourceRecordSetsWithContext(ctx, in)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured deletion of final zone delegation from intermediate zone")
	}

	return nil
}
