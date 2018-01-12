package endpointsv2

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the master endpoints in the Kubernetes API")

	namespace := keyv2.ClusterNamespace(customObject)

	// Lookup the current state of the endpoints.
	var endpoints *apiv1.Endpoints
	{
		manifest, err := r.k8sClient.CoreV1().Endpoints(namespace).Get(masterEndpointsName, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find the master endpoints in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found the master endpoints in the Kubernetes API")
			endpoints = manifest
		}
	}

	return endpoints, nil
}
