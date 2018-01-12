package endpointsv2

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	endpointsToCreate, err := toEndpoints(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if endpointsToCreate != nil {
		r.logger.LogCtx(ctx, "debug", "creating Kubernetes endpoints")

		namespace := keyv2.ClusterNamespace(customObject)
		_, err = r.k8sClient.CoreV1().Endpoints(namespace).Create(endpointsToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "creating Kubernetes endpoints: created")
	} else {
		r.logger.LogCtx(ctx, "debug", "creating Kubernetes endpoints: already created")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoints, err := toEndpoints(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoints, err := toEndpoints(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var endpointsToCreate *apiv1.Endpoints
	if currentEndpoints == nil || desiredEndpoints.Name != currentEndpoints.Name {
		endpointsToCreate = desiredEndpoints
	}

	return endpointsToCreate, nil
}
