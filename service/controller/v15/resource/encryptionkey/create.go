package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	createInput, err := toEncryptionKeyState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if createInput.KeyName != "" {
		err := r.encrypter.CreateKey(ctx, customObject, createInput.KeyName)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key: created")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key: already created")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentKeyState, err := toEncryptionKeyState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredKeyState, err := toEncryptionKeyState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if currentKeyState.KeyName == "" || desiredKeyState.KeyName != currentKeyState.KeyName {
		return desiredKeyState, nil
	}

	return nil, nil
}
