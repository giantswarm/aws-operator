package s3object

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v13/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	createBucketState, err := toBucketObjectState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	for key, bucketObject := range createBucketState {
		if bucketObject.Key != "" {
			s3PutInput, err := toPutObjectInput(bucketObject)
			if err != nil {
				return microerror.Mask(err)
			}

			_, err = sc.AWSClient.S3.PutObject(&s3PutInput)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 object '%s': created", key))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 object '%s': already created", key))
		}
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentBucketState, err := toBucketObjectState(currentState)
	if err != nil {
		return s3.PutObjectInput{}, microerror.Mask(err)
	}

	desiredBucketState, err := toBucketObjectState(desiredState)
	if err != nil {
		return s3.PutObjectInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the s3 objects should be created")

	createState := map[string]BucketObjectState{}

	for key, bucketObject := range desiredBucketState {
		if _, ok := currentBucketState[key]; !ok {
			createState[key] = bucketObject
		} else {
			createState[key] = BucketObjectState{}
		}
	}

	return createState, nil
}
