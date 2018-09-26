package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v17/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.encrypter.EnsureDeletedEncryptionKey(ctx, customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
