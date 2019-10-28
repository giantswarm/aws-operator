package secretfinalizer

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// When the operator's resource implementations request the CR's finalizers to
	// be kept, their deletion logic is delayed. That implies that we should not
	// remove the secret finalizers here already, since certain resource
	// implementations may still require secrets to be available during their own
	// deletion logic execution in upcoming reconciliation loops.
	if finalizerskeptcontext.IsKept(ctx) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not removing secret finalizers")
		r.logger.LogCtx(ctx, "level", "debug", "message", "finalizers requested to be kept")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
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

		if containsString(secret.Finalizers, secretFinalizer) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("removing finalizer for secret %#q in namespace %#q", s.Name, s.Namespace))

			secret.Finalizers = filterString(secret.Finalizers, secretFinalizer)

			_, err := r.k8sClient.CoreV1().Secrets(s.Namespace).Update(secret)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("removed finalizer for secret %#q in namespace %#q", s.Name, s.Namespace))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finalizer already removed for secret %#q in namespace %#q", s.Name, s.Namespace))
		}
	}

	return nil
}

func filterString(list []string, match string) []string {
	var filtered []string

	for _, s := range list {
		if s != match {
			filtered = append(filtered, s)
		}
	}

	return filtered
}
