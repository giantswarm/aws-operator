package s3bucket

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/service/controller/v22/controllercontext"
)

const (
	// loopLimit is the maximum amount of delete actions we want to allow per
	// S3 bucket. Reason here is to execute resources fast and prevent
	// them blocking other resources for too long. In case a S3 bucket has more
	// than 3000 objects, we delete 3 batches of 1000 objects and leave the rest
	// for the next reconciliation loop.
	loopLimit = 3
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

	g, ctx := errgroup.WithContext(ctx)

	for _, b := range bucketsInput {
		bucketName := b.Name

		g.Go(func() error {
			if bucketName == "" {
				return nil
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting S3 bucket %q", bucketName))

			var bucketEmpty bool
			var count int
			for {
				i := &s3.ListObjectsV2Input{
					Bucket: aws.String(bucketName),
				}
				o, err := sc.AWSClient.S3.ListObjectsV2(i)
				if err != nil {
					return microerror.Mask(err)
				}
				if len(o.Contents) == 0 {
					bucketEmpty = true
					break
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %d objects", len(o.Contents)))

				for _, o := range o.Contents {
					i := &s3.DeleteObjectInput{
						Bucket: aws.String(bucketName),
						Key:    o.Key,
					}
					_, err := sc.AWSClient.S3.DeleteObject(i)
					if err != nil {
						return microerror.Mask(err)
					}
				}

				count++
				if count >= loopLimit {
					r.logger.LogCtx(ctx, "level", "debug", "message", "loop limit reached for S3 bucket deletion")

					r.logger.LogCtx(ctx, "level", "debug", "message", "canceling S3 bucket deletion")
					break
				}
			}

			if bucketEmpty {
				i := &s3.DeleteBucketInput{
					Bucket: aws.String(bucketName),
				}

				_, err = sc.AWSClient.S3.DeleteBucket(i)
				if err != nil {
					return microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted S3 bucket %q", bucketName))
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("bucket %q not empty", bucketName))

				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)

				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
			}

			return nil
		})
	}

	err = g.Wait()
	if IsBucketNotFound(err) {
		// Fall through.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
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
