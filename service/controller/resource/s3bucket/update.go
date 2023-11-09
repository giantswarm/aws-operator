package s3bucket

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/crud"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	// update bucket tags
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	currentBuckets, err := toBucketState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, bucket := range currentBuckets {
		if r.includeTags {
			tagSet, err := r.getS3BucketTags(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			i := &s3.PutBucketTaggingInput{
				Bucket: aws.String(bucket.Name),
				Tagging: &s3.Tagging{
					TagSet: tagSet,
				},
			}

			_, err = cc.Client.TenantCluster.AWS.S3.PutBucketTagging(i)
			if err != nil {
				return microerror.Mask(err)
			}
		}
		r.logger.Debugf(ctx, "S3 bucket %#q tags updated", bucket.Name)
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

// newUpdateChange returns all buckets in order to update tags on all buckets
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentBuckets, err := toBucketState(currentState)

	return currentBuckets, err
}
