package tccpencryption

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.encrypter.EnsureDeletedEncryptionKey(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
