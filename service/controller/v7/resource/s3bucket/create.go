package s3bucket

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v7/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	bucketInput, err := toBucketState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if bucketInput.Name != "" {
		r.logger.LogCtx(ctx, "debug", "creating S3 bucket")

		_, err = r.clients.S3.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketInput.Name),
		})
		if IsBucketAlreadyExists(err) || IsBucketAlreadyOwnedByYou(err) {
			// Fall through.
			return nil
		}
		if err != nil {
			return microerror.Mask(err)
		}

		_, err := r.clients.S3.PutBucketTagging(&s3.PutBucketTaggingInput{
			Bucket: aws.String(bucketInput.Name),
			Tagging: &s3.Tagging{
				TagSet: getS3BucketTags(customObject),
			},
		})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "creating S3 bucket: created")
	} else {
		r.logger.LogCtx(ctx, "debug", "creating S3 bucket: already created")
	}
	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentBucket, err := toBucketState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredBucket, err := toBucketState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if currentBucket.Name == "" || desiredBucket.Name != currentBucket.Name {
		return desiredBucket, nil
	}

	return BucketState{}, nil
}
