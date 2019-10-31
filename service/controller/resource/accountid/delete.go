package accountid

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	err := r.addAccountIDToContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
