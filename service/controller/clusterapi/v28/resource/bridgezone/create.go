package bridgezone

import (
	"context"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	finalZone := key.ClusterID(cr) + ".k8s." + baseDomain

	guest, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	g := &errgroup.Group{}

	var intermediateZoneID string
	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting intermediate zone ID")

		id, err := r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone not found")

			return microerror.Mask(err)
		} else if err != nil {
			return microerror.Mask(err)
		}
		intermediateZoneID = id

		r.logger.LogCtx(ctx, "level", "debug", "message", "got intermediate zone ID")

		return nil
	})

	var finalZoneID string
	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone ID")

		id, err := r.findHostedZoneID(ctx, guest, finalZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "final zone not found")

			return microerror.Mask(err)
		} else if err != nil {
			return microerror.Mask(err)
		}
		finalZoneID = id

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone ID")

		return nil
	})

	err = g.Wait()
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var finalZoneRecords []*route53.ResourceRecord
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone name servers")

		nameServers, _, err := r.getNameServersAndTTL(ctx, guest, finalZoneID, finalZone)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, ns := range nameServers {
			copy := ns
			v := &route53.ResourceRecord{
				Value: &copy,
			}
			finalZoneRecords = append(finalZoneRecords, v)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone name servers")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring final zone delegation from intermediate zone")

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

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured final zone delegation from intermediate zone")
	}

	return nil
}
