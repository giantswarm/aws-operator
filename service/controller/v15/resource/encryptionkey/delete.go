package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v15/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	controllerCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var masterRoleARN string
	var workerRoleARN string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding master and worker role ARNs")

		accountID, err := controllerCtx.AWSService.GetAccountID()
		if err != nil {
			return microerror.Mask(err)
		}

		masterRoleARN = key.MasterRoleARN(customObject, accountID)
		workerRoleARN = key.WorkerRoleARN(customObject, accountID)

		r.logger.LogCtx(ctx, "level", "debug", "message", "found master and worker role ARNs")
	}

	{
		err = r.encrypter.EnsureDeletedAuthorizedIAMRoles(ctx, masterRoleARN, workerRoleARN)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of encryption key")

		current, err := r.encrypter.GetCurrentState(ctx, customObject)
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
