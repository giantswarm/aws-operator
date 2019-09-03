package endpoints

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var endpoints *corev1.Endpoints
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding endpoint")

		manifest, err := r.k8sClient.CoreV1().Endpoints(key.ClusterNamespace(cr)).Get(masterEndpointsName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find endpoint")
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found endpoint")
			endpoints = manifest
		}
	}

	return endpoints, nil
}
