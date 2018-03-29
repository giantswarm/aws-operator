package ebsvolume

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

// ApplyDeleteChange detaches and deletes the EBS volumes. We don't return
// errors so deletion in following resources is executed.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteInput, err := toEBSVolumeState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteInput != nil && len(deleteInput.Volumes) > 0 {
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("deleting %d ebs volumes", len(deleteInput.Volumes)))

		// First detach any attached volumes without forcing but shutdown the
		// instances.
		for _, vol := range deleteInput.Volumes {
			for _, a := range vol.Attachments {
				err := r.service.DetachVolume(ctx, vol.VolumeID, a, false, true)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("failed to detach EBS volume %s", vol.VolumeID), "stack", fmt.Sprintf("%#v", err))
				}
			}
		}

		// Now force detach so the volumes can be deleted cleanly.
		for _, vol := range deleteInput.Volumes {
			for _, a := range vol.Attachments {
				r.service.DetachVolume(ctx, vol.VolumeID, a, true, false)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("failed to force detach EBS volume %s", vol.VolumeID), "stack", fmt.Sprintf("%#v", err))
				}
			}
		}

		// Now delete the volumes.
		for _, vol := range deleteInput.Volumes {
			err := r.service.DeleteVolume(ctx, vol.VolumeID)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("failed to delete EBS volume %s", vol.VolumeID), "stack", fmt.Sprintf("%#v", err))
			}
		}

		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("deleted %d ebs volumes", len(deleteInput.Volumes)))
	} else {
		r.logger.LogCtx(ctx, "level", "info", "message", "not deleting load ebs volumes because there aren't any")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentVolState, err := toEBSVolumeState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredVolState, err := toEBSVolumeState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var volStateToDelete *EBSVolumeState
	if desiredVolState == nil && currentVolState != nil && len(currentVolState.Volumes) > 0 {
		volStateToDelete = currentVolState
	}

	return volStateToDelete, nil
}
