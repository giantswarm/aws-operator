package kmskey

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"

	servicecontext "github.com/giantswarm/aws-operator/service/controller/v11/context"
	"github.com/giantswarm/aws-operator/service/controller/v11/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	createInput, err := toKMSKeyState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if createInput.KeyAlias != "" {
		sc, err := servicecontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		key, err := sc.AWSClient.KMS.CreateKey(&kms.CreateKeyInput{})
		if err != nil {
			return microerror.Mask(err)
		}

		if _, err := sc.AWSClient.KMS.CreateAlias(&kms.CreateAliasInput{
			AliasName:   aws.String(createInput.KeyAlias),
			TargetKeyId: key.KeyMetadata.Arn,
		}); err != nil {
			return microerror.Mask(err)
		}

		if _, err := sc.AWSClient.KMS.EnableKeyRotation(&kms.EnableKeyRotationInput{
			KeyId: key.KeyMetadata.KeyId,
		}); err != nil {
			return microerror.Mask(err)
		}

		if _, err := sc.AWSClient.KMS.TagResource(&kms.TagResourceInput{
			KeyId: key.KeyMetadata.KeyId,
			Tags:  r.getKMSTags(customObject),
		}); err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "creating KMS key: created")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating KMS key: already created")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentKeyState, err := toKMSKeyState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredKeyState, err := toKMSKeyState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if currentKeyState.KeyAlias == "" || desiredKeyState.KeyAlias != currentKeyState.KeyAlias {
		return desiredKeyState, nil
	}

	return KMSKeyState{}, nil
}
