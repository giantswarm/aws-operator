package endpointsv2

import (
	"context"
	"reflect"

	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	endpointsToUpdate, err := toEndpoints(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if endpointsToUpdate != nil {
		r.logger.LogCtx(ctx, "debug", "updating Kubernetes endpoints")

		namespace := keyv2.ClusterNamespace(customObject)
		_, err := r.k8sClient.CoreV1().Endpoints(namespace).Update(endpointsToUpdate)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "updated Kubernetes endpoints")

	} else {
		r.logger.LogCtx(ctx, "debug", "Kubernetes endpoints do not need to be updated")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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

	r.logger.LogCtx(ctx, "debug", "finding out if the endpoints has to be updated")

	var endpointsToUpdate *apiv1.Endpoints

	// The subsets can change if the private IP of the master node has changed.
	// We then need to update the endpoints resource.
	if currentEndpoints != nil && desiredEndpoints != nil {
		if !reflect.DeepEqual(desiredEndpoints.Subsets, currentEndpoints.Subsets) {
			endpointsToUpdate = desiredEndpoints
		}
	}

	r.logger.LogCtx(ctx, "debug", "found out if the endpoints has to be deleted")

	return endpointsToUpdate, nil
}
