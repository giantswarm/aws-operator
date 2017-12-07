package kmskeyv1

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	createKeyInput, err := toCreateKeyInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if createKeyInput != nil {
		customObject, err := keyv1.ToCustomObject(obj)
		if err != nil {
			return microerror.Mask(err)
		}
		key, err := r.awsClients.KMS.CreateKey(createKeyInput)
		if err != nil {
			return microerror.Mask(err)
		}

		clusterID := keyv1.ClusterID(customObject)
		keyAlias := toAlias(clusterID)
		if _, err := r.awsClients.KMS.CreateAlias(&kms.CreateAliasInput{
			AliasName:   aws.String(keyAlias),
			TargetKeyId: key.KeyMetadata.Arn,
		}); err != nil {
			return microerror.Mask(err)
		}

		if _, err := r.awsClients.KMS.EnableKeyRotation(&kms.EnableKeyRotationInput{
			KeyId: key.KeyMetadata.KeyId,
		}); err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "creating KMS key: created")
	} else {
		r.logger.LogCtx(ctx, "debug", "creating KMS key: already created")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentKMSState, err := toKMSKeyState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredKMSState, err := toKMSKeyState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if desiredKMSState.KeyID != currentKMSState.KeyID {
		return &kms.CreateKeyInput{}, nil
	}

	return nil, nil
}
