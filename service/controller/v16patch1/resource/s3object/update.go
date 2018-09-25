package s3object

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/aws-operator/service/controller/v16patch1/controllercontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateBucketState, err := toBucketObjectState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	for key, bucketObject := range updateBucketState {
		if bucketObject.Key != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating S3 object '%s'", key))

			s3PutInput, err := toPutObjectInput(bucketObject)
			if err != nil {
				return microerror.Mask(err)
			}

			_, err = sc.AWSClient.S3.PutObject(&s3PutInput)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated S3 object '%s'", key))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not update S3 object '%s'", key))
			r.logger.LogCtx(ctx, "level", "debug", "message", "S3 object already up to date")
		}
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentS3Object, err := toBucketObjectState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredS3Object, err := toBucketObjectState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the s3 objects should be updated")

	updateState := map[string]BucketObjectState{}

	for key, bucketObject := range desiredS3Object {
		if _, ok := currentS3Object[key]; !ok {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object '%s' should not be updated", key))
			updateState[key] = BucketObjectState{}
		}

		currentObject := currentS3Object[key]
		if currentObject.Body != "" && bucketObject.Body != currentObject.Body {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object '%s' should be updated", key))
			updateState[key] = bucketObject
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object '%s' should not be updated", key))
			updateState[key] = BucketObjectState{}
		}
	}

	return updateState, nil
}
