package s3bucket

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/aws-operator/service/controller/v14patch3/controllercontext"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	bucketsInput, err := toBucketState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, bucketInput := range bucketsInput {
		if bucketInput.Name == "" {
			continue
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting S3 bucket %q", bucketInput.Name))

		var repeat bool
		for {
			i := &s3.ListObjectsV2Input{
				Bucket: aws.String(bucketInput.Name),
			}
			o, err := sc.AWSClient.S3.ListObjectsV2(i)
			if err != nil {
				return microerror.Mask(err)
			}
			if o.IsTruncated != nil && *o.IsTruncated {
				repeat = true
			}
			if len(o.Contents) == 0 {
				break
			}

			for _, o := range o.Contents {
				i := &s3.DeleteObjectInput{
					Bucket: aws.String(bucketInput.Name),
					Key:    o.Key,
				}
				_, err := sc.AWSClient.S3.DeleteObject(i)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			if !repeat {
				break
			}
		}

		_, err = sc.AWSClient.S3.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketInput.Name),
		})
		if IsBucketNotFound(err) {
			// Fall through.
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting S3 bucket %q: deleted", bucketInput.Name))
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
	currentBuckets, err := toBucketState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredBuckets, err := toBucketState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var bucketsToDelete []BucketState
	for _, bucket := range currentBuckets {
		// Destination Logs Bucket should not be deleted because it has to keep logs
		// even when cluster is removed (rotation of these logs are managed externally).
		if r.canBeDeleted(bucket) && containsBucketState(bucket.Name, desiredBuckets) {
			bucketsToDelete = append(bucketsToDelete, bucket)
		}
	}

	return bucketsToDelete, nil
}
