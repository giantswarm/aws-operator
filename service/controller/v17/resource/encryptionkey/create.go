package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v17/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
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
		err = r.encrypter.EnsureCreatedAuthorizedIAMRoles(ctx, masterRoleARN, workerRoleARN)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring encryption key")

		current, err := r.encrypter.GetCurrentState(ctx, customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		desired, err := r.encrypter.GetDesiredState(ctx, customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		if current.KeyName == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key")

			err = r.encrypter.CreateKey(ctx, customObject, desired.KeyName)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "created encryption key")
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "encryption key already created")
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured encryption key")
	}

	return nil
}
