package cphostedzone

import (
	"context"

	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	err = r.addHostedZoneInfoToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
