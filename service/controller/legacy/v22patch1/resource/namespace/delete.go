package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	namespaceToDelete, err := toNamespace(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToDelete != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting Kubernetes namespace")

		err = r.k8sClient.CoreV1().Namespaces().Delete(namespaceToDelete.Name, &apismetav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting Kubernetes namespace: deleted")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting Kubernetes namespace: already deleted")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := toNamespace(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the namespace has to be deleted")

	var namespaceToDelete *apiv1.Namespace
	if currentNamespace != nil {
		namespaceToDelete = desiredNamespace
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found out if the namespace has to be deleted")

	return namespaceToDelete, nil
}
