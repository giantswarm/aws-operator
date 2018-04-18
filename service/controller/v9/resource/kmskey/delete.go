package kmskey

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteInput, err := toKMSKeyState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteInput.KeyAlias != "" {
		// Get the KMS key ID using the key alias.
		key, err := r.awsClients.KMS.DescribeKey(&kms.DescribeKeyInput{
			KeyId: aws.String(deleteInput.KeyAlias),
		})
		if err != nil {
			return microerror.Mask(err)
		}

		// Delete the key alias.
		if _, err := r.awsClients.KMS.DeleteAlias(&kms.DeleteAliasInput{
			AliasName: aws.String(deleteInput.KeyAlias),
		}); err != nil {
			return microerror.Mask(err)
		}

		// AWS API doesn't allow to delete the KMS key immediately, but we can schedule its deletion.
		if _, err := r.awsClients.KMS.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
			KeyId:               key.KeyMetadata.KeyId,
			PendingWindowInDays: aws.Int64(pendingDeletionWindow),
		}); err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting KMS Key: deleted")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting KMS Key: already deleted")
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
	currentKeyState, err := toKMSKeyState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredKeyState, err := toKMSKeyState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the KMS key should be deleted")

	var kmsKeyToDelete KMSKeyState
	if currentKeyState.KeyAlias != "" {
		kmsKeyToDelete = desiredKeyState
	}

	return kmsKeyToDelete, nil
}
