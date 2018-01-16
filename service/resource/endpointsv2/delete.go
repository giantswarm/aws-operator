package endpointsv2

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	endpointsToDelete, err := toEndpoints(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if endpointsToDelete != nil {
		r.logger.LogCtx(ctx, "debug", "deleting Kubernetes endpoints")

		namespace := keyv2.ClusterNamespace(customObject)
		err := r.k8sClient.CoreV1().Endpoints(namespace).Delete(endpointsToDelete.Name, &apismetav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "deleting Kubernetes endpoints: deleted")
	} else {
		r.logger.LogCtx(ctx, "debug", "deleting Kubernetes endpoints: already deleted")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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

	r.logger.LogCtx(ctx, "debug", "finding out if the endpoints has to be deleted")

	var endpointsToDelete *apiv1.Endpoints
	if currentEndpoints != nil && desiredEndpoints.Name == currentEndpoints.Name {
		endpointsToDelete = desiredEndpoints
	}

	r.logger.LogCtx(ctx, "debug", "found out if the endpoints has to be deleted")

	return endpointsToDelete, nil
}
