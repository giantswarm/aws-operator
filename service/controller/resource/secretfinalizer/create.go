package secretfinalizer

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v2/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, s := range newSecretAccessors(ctx, cr) {
		var secret *corev1.Secret
		{
			r.logger.Debugf(ctx, "finding secret %#q in namespace %#q", s.Name, s.Namespace)

			secret, err = r.k8sClient.CoreV1().Secrets(s.Namespace).Get(ctx, s.Name, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				r.logger.Debugf(ctx, "did not find secret %#q in namespace %#q", s.Name, s.Namespace)
				r.logger.Debugf(ctx, "continuing with next secret")
				continue

			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "found secret %#q in namespace %#q", s.Name, s.Namespace)
		}

		if !containsString(secret.Finalizers, secretFinalizer) {
			r.logger.Debugf(ctx, "adding finalizer for secret %#q in namespace %#q", s.Name, s.Namespace)

			secret.Finalizers = append(secret.Finalizers, secretFinalizer)

			_, err := r.k8sClient.CoreV1().Secrets(s.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "added finalizer for secret %#q in namespace %#q", s.Name, s.Namespace)
		} else {
			r.logger.Debugf(ctx, "finalizer already added for secret %#q in namespace %#q", s.Name, s.Namespace)
		}
	}

	return nil
}

func containsString(list []string, match string) bool {
	for _, s := range list {
		if s == match {
			return true
		}
	}

	return false
}
