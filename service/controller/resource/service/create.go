package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	serviceToCreate, err := toService(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if serviceToCreate != nil {
		r.logger.Debugf(ctx, "creating service")

		namespace := key.ClusterNamespace(cr)
		_, err = r.k8sClient.CoreV1().Services(namespace).Create(ctx, serviceToCreate, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not create service")
			r.logger.Debugf(ctx, "namespace not found yet")
			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "created service")
	} else {
		r.logger.Debugf(ctx, "did not create service")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentService, err := toService(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredService, err := toService(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var serviceToCreate *corev1.Service
	if currentService == nil || desiredService.Name != currentService.Name {
		serviceToCreate = desiredService
	}

	return serviceToCreate, nil
}
