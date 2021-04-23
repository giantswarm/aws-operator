package tccpsubnets

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	{
		r.logger.Debugf(ctx, "finding the tenant cluster's control plane subnets")

		err := r.addSubnetsToContext(ctx)
		if IsVPCNotFound(err) {
			r.logger.Debugf(ctx, "the tenant cluster's control plane vpc id is not available yet")
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found the tenant cluster's control plane subnets")
	}

	return nil
}
