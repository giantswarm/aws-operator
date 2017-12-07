package kmskeyv1

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteAliasInput, err := toDeleteAliasInput(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	keyAlias := deleteAliasInput.AliasName
	if keyAlias != nil && *keyAlias != "" {
		if _, err := r.awsClients.KMS.DeleteAlias(&deleteAliasInput); err != nil {
			return microerror.Mask(err)
		}

		key, err := r.awsClients.KMS.DescribeKey(&kms.DescribeKeyInput{
			KeyId: keyAlias,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		// AWS API doesn't allow to delete the KMS key immediately, but we can schedule its deletion
		if _, err := r.awsClients.KMS.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
			KeyId:               key.KeyMetadata.KeyId,
			PendingWindowInDays: aws.Int64(pendingDeletionWindow),
		}); err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "deleting KMS Key: deleted")
	} else {
		r.logger.LogCtx(ctx, "debug", "deleting KMS Key: already deleted")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	deleteChange := kms.DeleteAliasInput{}
	currentKeyState, err := toKMSKeyState(currentState)
	if err != nil {
		return deleteChange, microerror.Mask(err)
	}

	desiredKeyState, err := toKMSKeyState(desiredState)
	if err != nil {
		return deleteChange, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the KMS key should be deleted")

	if currentKeyState.KeyAlias != "" && desiredKeyState.KeyAlias != currentKeyState.KeyAlias {
		deleteChange.AliasName = aws.String(currentKeyState.KeyAlias)
	}

	return deleteChange, nil
}
