package tccpsubnets

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane subnets")

		err := r.addSubnetsToContext(ctx)
		if IsVPCNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane vpc id is not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane subnets")
	}

	return nil
}
