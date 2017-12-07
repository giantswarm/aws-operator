package s3objectv2

import (
	"context"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateObjectInput, err := toPutObjectInput(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	objectName := updateObjectInput.Key
	if *objectName != "" {
		_, err := r.awsClients.S3.PutObject(&updateObjectInput)
		if err != nil {
			return microerror.Maskf(err, "updating S3 object")
		}

		r.logger.LogCtx(ctx, "debug", "updating S3 object: updated")
	} else {
		r.logger.LogCtx(ctx, "debug", "updating S3 object: no need to update")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	output := s3.PutObjectInput{}

	desiredObjectState, err := toBucketObjectState(desiredState)
	if err != nil {
		return output, microerror.Mask(err)
	}

	currentObjectState, err := toBucketObjectState(currentState)
	if err != nil {
		return output, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if S3 object should be updated")

	if !reflect.DeepEqual(desiredObjectState, currentObjectState) {
		output.Key = aws.String(desiredObjectState.WorkerCloudConfig.Key)
		output.Body = strings.NewReader(desiredObjectState.WorkerCloudConfig.Body)
		output.Bucket = aws.String(desiredObjectState.WorkerCloudConfig.Bucket)
		output.ContentLength = aws.Int64(int64(len(desiredObjectState.WorkerCloudConfig.Body)))
	}

	return output, nil
}
