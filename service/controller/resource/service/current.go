package service

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v16/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for the service in the Kubernetes API")

	namespace := key.ClusterNamespace(cr)

	// Lookup the current state of the service.
	var service *corev1.Service
	{
		manifest, err := r.k8sClient.CoreV1().Services(namespace).Get(ctx, masterServiceName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find the service in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found the service in the Kubernetes API")
			service = manifest
		}
	}

	return service, nil
}
