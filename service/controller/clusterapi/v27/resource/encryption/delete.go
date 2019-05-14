package encryption

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.encrypter.EnsureDeletedEncryptionKey(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
