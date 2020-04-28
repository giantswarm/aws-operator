package s3object

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateBucketState, err := toBucketObjectState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	for objectKey, bucketObject := range updateBucketState {
		if bucketObject.Key != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating S3 object %#q", objectKey))

			s3PutInput, err := toPutObjectInput(bucketObject)
			if err != nil {
				return microerror.Mask(err)
			}

			_, err = cc.Client.TenantCluster.AWS.S3.PutObject(&s3PutInput)
			if err != nil {
				return microerror.Mask(err)
			}

			switch objectKey {
			case key.BucketObjectName(key.KindMaster):
				cc.Spec.TenantCluster.MasterInstance.IgnitionHash = bucketObject.Hash
			case key.BucketObjectName(key.KindWorker):
				cc.Spec.TenantCluster.WorkerInstance.IgnitionHash = bucketObject.Hash
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated S3 object %#q", objectKey))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not update S3 object %#q", objectKey))
		}
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
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object %#q should not be updated", key))
			updateState[key] = BucketObjectState{}
		}

		currentObject := currentS3Object[key]
		if currentObject.Body != "" && bucketObject.Body != currentObject.Body {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object %#q should be updated", key))
			updateState[key] = bucketObject
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("S3 object %#q should not be updated", key))
			updateState[key] = BucketObjectState{}
		}
	}

	return updateState, nil
}
