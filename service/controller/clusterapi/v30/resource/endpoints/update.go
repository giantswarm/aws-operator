package endpoints

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	endpointsToUpdate, err := toEndpoints(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if endpointsToUpdate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating endpoint")

		namespace := key.ClusterNamespace(cr)
		_, err := r.k8sClient.CoreV1().Endpoints(namespace).Update(endpointsToUpdate)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated endpoint")

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not update endpoint")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoints, err := toEndpoints(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoints, err := toEndpoints(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var endpointsToUpdate *corev1.Endpoints

	// The subsets can change if the private IP of the master node has changed.
	// We then need to update the endpoints resource.
	if currentEndpoints != nil && desiredEndpoints != nil {
		if !reflect.DeepEqual(desiredEndpoints.Subsets, currentEndpoints.Subsets) {
			endpointsToUpdate = desiredEndpoints
		}
	}

	return endpointsToUpdate, nil
}
