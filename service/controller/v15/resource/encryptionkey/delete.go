package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/aws-operator/service/controller/v15/encrypter"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteInput, err := toEncryptionKeyState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteInput.KeyName != "" {
		err := r.encrypter.DeleteKey(ctx, deleteInput.KeyName)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryption Key: deleted")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryption Key: already deleted")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentKeyState, err := toEncryptionKeyState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredKeyState, err := toEncryptionKeyState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the encryption key should be deleted")

	var keyToDelete encrypter.EncryptionKeyState
	if currentKeyState.KeyName != "" {
		keyToDelete = desiredKeyState
	}

	return keyToDelete, nil
}
