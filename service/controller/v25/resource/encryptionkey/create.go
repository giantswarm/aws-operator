package encryptionkey

import (
	"context"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.encrypter.EnsureCreatedEncryptionKey(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	var encryptionKey string
	{
		o := func() error {
			encryptionKey, err = r.encrypter.EncryptionKey(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	cc.Status.TenantCluster.EncryptionKey = encryptionKey

	return nil
}
