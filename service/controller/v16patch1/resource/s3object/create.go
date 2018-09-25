package s3object

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v16patch1/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	s3ObjectToCreate, err := toBucketObjectState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	for key, bucketObject := range s3ObjectToCreate {
		if bucketObject.Key != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 object '%s'", key))

			s3PutInput, err := toPutObjectInput(bucketObject)
			if err != nil {
				return microerror.Mask(err)
			}

			_, err = sc.AWSClient.S3.PutObject(&s3PutInput)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 object '%s'", key))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not create S3 object '%s'", key))
			r.logger.LogCtx(ctx, "level", "debug", "message", "S3 object already exists")
		}
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentS3Object, err := toBucketObjectState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredS3Object, err := toBucketObjectState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the s3 objects should be created")

	createState := map[string]BucketObjectState{}

	for key, bucketObject := range desiredS3Object {
		_, ok := currentS3Object[key]
		if !ok {
			// The desired object does not exist in the current state of the system,
			// so we want to create it.
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object '%s' should be created", key))
			createState[key] = bucketObject
		} else {
			// The desired object exists in the current state of the system, so we do
			// not want to create it. We do track it using an empty object reference
			// though, in order to get some more useful logging in ApplyCreateChange.
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object '%s' should not be created", key))
			createState[key] = BucketObjectState{}
		}
	}

	return createState, nil
}
