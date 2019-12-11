package tcnpencryption

import (
	"context"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var cl infrastructurev1alpha2.AWSCluster
	{
		md, err := key.ToMachineDeployment(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		n := types.NamespacedName{
			Name:      key.ClusterID(&md),
			Namespace: md.Namespace,
		}
		m, err := r.g8sClient.InfrastructureV1alpha2().AWSClusters().Get(n.String(), metav1.GetOptions{})
		if errors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not yet availabile")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		cl = *m
	}

	// For some obscure reasons the encryption key is not immediately available
	// when creating it. On each cluster creation we saw the retry resource
	// kicking in once because of a not found error. To prevent the error, instead
	// we backoff silently upfront where we know we have to.
	{
		var encryptionKey string

		o := func() error {
			encryptionKey, err = r.encrypter.EncryptionKey(ctx, cl)
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
