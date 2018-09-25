package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v16patch1/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of encryption key")

		current, err := r.encrypter.CurrentState(ctx, customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		if current.KeyName != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryption key")

			err = r.encrypter.DeleteKey(ctx, current.KeyName)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "deleted encryption key")
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "encryption key already deleted")
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured deletion of encryption key")
	}

	return nil
}
