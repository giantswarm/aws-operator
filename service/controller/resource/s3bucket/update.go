package s3bucket

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v2/pkg/resource/crud"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
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

// newUpdateChange is a no-op because S3 buckets are not updated.
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}
