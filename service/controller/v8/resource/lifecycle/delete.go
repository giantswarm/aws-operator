package lifecycle

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	return nil, nil
}
