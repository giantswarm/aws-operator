package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	serviceToUpdate, err := toService(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if serviceToUpdate != nil && serviceToUpdate.Spec.ClusterIP != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating services")

		_, err := r.k8sClient.CoreV1().Services(serviceToUpdate.Namespace).Update(ctx, serviceToUpdate, metav1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated services")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not update service")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

// Service resources are updated.
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentService, err := toService(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredService, err := toService(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the service has to be updated")

	if isServiceModified(desiredService, currentService) {
		// Make a copy and set the resource version so the service can be updated.
		serviceToUpdate := desiredService.DeepCopy()
		if currentService != nil {
			serviceToUpdate.ObjectMeta.ResourceVersion = currentService.ObjectMeta.ResourceVersion
			serviceToUpdate.Spec.ClusterIP = currentService.Spec.ClusterIP
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "the service has to be updated")

		return serviceToUpdate, nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "the service does not have to be updated")

	return nil, nil
}
