package service

import (
	"context"

	"fmt"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	serviceToUpdate, err := toService(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if serviceToUpdate != nil && serviceToUpdate.Spec.ClusterIP != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating services")

		_, err := r.k8sClient.CoreV1().Services(serviceToUpdate.Namespace).Update(serviceToUpdate)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated services")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to update services")
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

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which services have to be updated")

	if isServiceModified(desiredService, currentService) {
		// Make a copy and set the resource version so the service can be updated.
		serviceToUpdate := desiredService.DeepCopy()
		if currentService != nil {
			serviceToUpdate.ObjectMeta.ResourceVersion = currentService.ObjectMeta.ResourceVersion
			serviceToUpdate.Spec.ClusterIP = currentService.Spec.ClusterIP
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found service '%s' that has to be updated", desiredService.GetName()))

		return serviceToUpdate, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no services needs update")

		return nil, nil
	}
}
