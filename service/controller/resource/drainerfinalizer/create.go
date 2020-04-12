package drainerfinalizer

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.ensure(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
