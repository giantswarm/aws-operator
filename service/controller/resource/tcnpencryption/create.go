package tcnpencryption

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

		m, err := r.g8sClient.InfrastructureV1alpha2().AWSClusters(md.Namespace).Get(key.ClusterID(&md), metav1.GetOptions{})
		if errors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		cl = *m
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
