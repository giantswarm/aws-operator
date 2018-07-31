package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring encryption key")

		current, err := r.encrypter.CurrentState(ctx, customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		desired, err := r.encrypter.DesiredState(ctx, customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		if current.KeyName == "" {
			err = r.encrypter.CreateKey(ctx, customObject, desired.KeyName)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured encryption key")
	}

	return nil
}
