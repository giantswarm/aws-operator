package tccpdetachlbsubnet

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.ensureUnusedAZsAreDetachedFromLBs(ctx, obj)
	if IsVPCNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane vpc id is not yet available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
