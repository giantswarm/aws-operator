package s3bucketv1

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	bucketInput, err := toBucketState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if bucketInput.Name != "" {
		r.logger.LogCtx(ctx, "debug", "deleting S3 bucket")

		_, err := r.clients.S3.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketInput.Name),
		})
		if IsBucketNotFound(err) {
			// Fall through.
			return nil
		}
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "deleting S3 bucket: deleted")
	} else {
		r.logger.LogCtx(ctx, "debug", "deleting S3 bucket: already deleted")
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
	currentBucket, err := toBucketState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredBucket, err := toBucketState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var bucketToDelete BucketState
	if currentBucket.Name != "" {
		bucketToDelete = desiredBucket
	}

	return bucketToDelete, nil
}
