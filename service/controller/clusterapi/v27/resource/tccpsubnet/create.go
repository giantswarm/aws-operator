package tccpsubnet

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := legacykey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's tccp subnets")

		err = r.addSubnetsToContext(ctx, cr)
		if IsVPCNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's tccp vpc")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's tccp subnets")
	}

	return nil
}
