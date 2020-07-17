package secretfinalizer

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, s := range newSecretAccessors(ctx, cr) {
		var secret *corev1.Secret
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding secret %#q in namespace %#q", s.Name, s.Namespace))

			secret, err = r.k8sClient.CoreV1().Secrets(s.Namespace).Get(s.Name, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find secret %#q in namespace %#q", s.Name, s.Namespace))
				r.logger.LogCtx(ctx, "level", "debug", "message", "continuing with next secret")
				continue

			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found secret %#q in namespace %#q", s.Name, s.Namespace))
		}

		if !containsString(secret.Finalizers, secretFinalizer) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("adding finalizer for secret %#q in namespace %#q", s.Name, s.Namespace))

			secret.Finalizers = append(secret.Finalizers, secretFinalizer)

			_, err := r.k8sClient.CoreV1().Secrets(s.Namespace).Update(secret)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("added finalizer for secret %#q in namespace %#q", s.Name, s.Namespace))
			r.event.Emit(ctx, &cr, "FinalizerSecretCreated", fmt.Sprintf("Added finalizer for secret %#q in namespace %#q", s.Name, s.Namespace))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finalizer already added for secret %#q in namespace %#q", s.Name, s.Namespace))
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
