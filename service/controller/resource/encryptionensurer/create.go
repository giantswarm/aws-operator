package encryptionensurer

import (
	"context"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// The tccpencryption resource is reconciled upon the AWSCluster CR which
	// defines the TCCP Cloud Formation stack. With the provider specific Cluster
	// CR we ensure the encryption key. Note that the TCNP stack for the Node
	// Pools also needs the encryption key.
	err = r.encrypter.EnsureCreatedEncryptionKey(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	// For some obscure reasons the encryption key is not immediately available
	// when creating it. On each cluster creation we saw the retry resource
	// kicking in once because of a not found error. To prevent the error, instead
	// we backoff silently upfront where we know we have to.
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
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.Encryption.Key = encryptionKey
	}

	return nil
}
