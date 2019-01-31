package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v22/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v22/key"
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

	encryptionKey, err := r.encrypter.EncryptionKey(ctx, customObject)
	if r.encrypter.IsKeyNotFound(err) {
		// We can get here during deletion, if the key is already deleted we can
		// safely exit.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	cc.Status.Cluster.EncryptionKey = encryptionKey

	return nil
}
