package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/aws-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// No-op if we are not using cloudformation.
	if !key.UseCloudFormation(customObject) {
		r.logger.LogCtx(ctx, "debug", "not processing Kubernetes namespace")
		return nil, nil
	}

	r.logger.LogCtx(ctx, "debug", "looking for the namespace in the Kubernetes API")

	// Lookup the current state of the namespace.
	var namespace *apiv1.Namespace
	{
		manifest, err := r.k8sClient.CoreV1().Namespaces().Get(key.ClusterNamespace(customObject), apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find the namespace in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found the namespace in the Kubernetes API")
			namespace = manifest
		}
	}

	return namespace, nil
}
