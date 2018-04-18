package ebsvolume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteInput, err := toEBSVolumeState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteInput != nil && len(deleteInput.VolumeIDs) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %d ebs volumes", len(deleteInput.VolumeIDs)))

		for _, volID := range deleteInput.VolumeIDs {
			_, err := r.clients.EC2.DeleteVolume(&ec2.DeleteVolumeInput{
				VolumeId: aws.String(volID),
			})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted %d ebs volumes", len(deleteInput.VolumeIDs)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting load ebs volumes because there aren't any")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
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
	if desiredVolState == nil && currentVolState != nil && len(currentVolState.VolumeIDs) > 0 {
		volStateToDelete = currentVolState
	}

	return volStateToDelete, nil
}
