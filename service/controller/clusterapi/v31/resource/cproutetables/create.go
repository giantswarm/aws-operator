package cproutetables

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.addRouteTablesToContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
