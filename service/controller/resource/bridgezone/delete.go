package bridgezone

import (
	"context"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v12/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.Debugf(ctx, "route53 disabled")
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	baseDomain := key.ClusterBaseDomain(cr)
	intermediateZone := "k8s." + baseDomain
	finalZone := key.ClusterID(&cr) + ".k8s." + baseDomain

	_, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID string
	{
		r.logger.Debugf(ctx, "getting intermediate zone ID")

		intermediateZoneID, err = r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			r.logger.Debugf(ctx, "intermediate zone not found")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "got intermediate zone ID")
	}

	var finalZoneTTL int64
	var finalZoneRecords []*route53.ResourceRecord
	{
		r.logger.Debugf(ctx, "getting final zone delegation name servers and TTL from intermediate zone")

		nameServers, ttl, err := r.getNameServersAndTTL(ctx, defaultGuest, intermediateZoneID, finalZone)
		if IsNotFound(err) {
			// Delegation may be already deleted. It must be handled.
			r.logger.Debugf(ctx, "final zone delegation not found in intermediate zone")
			r.logger.Debugf(ctx, "canceling resource")
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

		r.logger.Debugf(ctx, "got final zone delegation name servers and TTL from intermediate zone")
	}

	{
		r.logger.Debugf(ctx, "ensuring deletion of final zone delegation from intermediate zone")

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

		r.logger.Debugf(ctx, "ensured deletion of final zone delegation from intermediate zone")
	}

	return nil
}
