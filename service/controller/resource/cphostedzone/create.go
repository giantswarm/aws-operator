package cphostedzone

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.Debugf(ctx, "route53 disabled")
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	err = r.addHostedZoneInfoToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
