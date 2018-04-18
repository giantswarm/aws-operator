package service

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/aws-operator/service/awsconfig/v7/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	serviceToCreate, err := toService(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if serviceToCreate != nil {
		r.logger.LogCtx(ctx, "debug", "creating Kubernetes service")

		namespace := key.ClusterNamespace(customObject)
		_, err = r.k8sClient.CoreV1().Services(namespace).Create(serviceToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "creating Kubernetes service: created")
	} else {
		r.logger.LogCtx(ctx, "debug", "creating Kubernetes service: already created")
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

	var serviceToCreate *apiv1.Service
	if currentService == nil || desiredService.Name != currentService.Name {
		serviceToCreate = desiredService
	}

	return serviceToCreate, nil
}
