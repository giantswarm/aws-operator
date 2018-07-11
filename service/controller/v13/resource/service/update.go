package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apiv1 "k8s.io/api/core/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	servicesToUpdate, err := toServices(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(servicesToUpdate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating services")

		for _, serviceToUpdate := range servicesToUpdate {
			_, err := r.k8sClient.CoreV1().Services(serviceToUpdate.Namespace).Update(serviceToUpdate)
			if err != nil {
				return microerror.Mask(err)
			}
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
	currentServices, err := toServices(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServices, err := toServices(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which services have to be updated")

	servicesToUpdate := make([]*apiv1.Service, 0)

	for _, currentService := range currentServices {
		desiredService, err := getServiceByName(desiredServices, currentService.Name)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isServiceModified(desiredService, currentService) {
			// Make a copy and set the resource version so the service can be updated.
			serviceToUpdate := desiredService.DeepCopy()
			serviceToUpdate.ObjectMeta.ResourceVersion = currentService.ObjectMeta.ResourceVersion

			servicesToUpdate = append(servicesToUpdate, serviceToUpdate)

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found service '%s' that has to be updated", desiredService.GetName()))
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d services which have to be updated", len(servicesToUpdate)))

	return servicesToUpdate, nil
}
