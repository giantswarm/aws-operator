package encryptionsearcher

import (
	"context"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// We need to wait until the encryption key got created. So in case
	// we do not find it, we cancel the resource and try again during the next
	// reconciliation loop.
	{
		var encryptionKey string

		o := func() error {
			encryptionKey, err = r.encrypter.EncryptionKey(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)

		err := backoff.Retry(o, b)
		if r.encrypter.IsKeyNotFound(err) {
			r.logger.Debugf(ctx, "encryption key not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.Encryption.Key = encryptionKey
	}

	return nil
}
