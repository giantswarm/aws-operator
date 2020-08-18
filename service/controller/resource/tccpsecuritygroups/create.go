package tccpsecuritygroups

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addInfoToCtx(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
