package kmskeyarn

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding tenant cluster kms key arn")

		err = r.addKMSKeyARNToContext(ctx, cr)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find tenant cluster kms key arn")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found tenant cluster kms key arn")
	}

	return nil
}
