package endpoints

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	endpointsToDelete, err := toEndpoints(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if endpointsToDelete != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting endpoint")

		namespace := key.ClusterNamespace(cr)
		err := r.k8sClient.CoreV1().Endpoints(namespace).Delete(ctx, endpointsToDelete.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted endpoint")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete endpoint")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoints, err := toEndpoints(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoints, err := toEndpoints(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var endpointsToDelete *corev1.Endpoints
	if currentEndpoints != nil && desiredEndpoints.Name == currentEndpoints.Name {
		endpointsToDelete = desiredEndpoints
	}

	return endpointsToDelete, nil
}
