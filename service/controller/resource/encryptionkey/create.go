package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cl, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// The encryption key is created within the cluster controller. This here is
	// the machine deployment controller. We need to wait until the encryption key
	// got created. So in case we do not find it, we cancel the resource and try
	// again during the next reconciliation loop.
	{
		encryptionKey, err := r.encrypter.EncryptionKey(ctx, cl)
		if r.encrypter.IsKeyNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "encryption key not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.Encryption.Key = encryptionKey
	}

	return nil
}
