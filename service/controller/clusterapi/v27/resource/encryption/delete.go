package encryption

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := legacykey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.encrypter.EnsureDeletedEncryptionKey(ctx, customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
