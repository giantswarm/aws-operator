package ebsvolume

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteInput, err := toEBSVolumeState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteInput != nil && len(deleteInput.Volumes) > 0 {
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("deleting %d ebs volumes", len(deleteInput.Volumes)))

		// First detach any attached volumes.
		for _, vol := range deleteInput.Volumes {
			for _, a := range vol.Attachments {
				r.detachVolume(ctx, vol.VolumeID, a)
			}
		}

		// Now delete the volumes.
		for _, vol := range deleteInput.Volumes {
			err := r.deleteVolume(ctx, vol.VolumeID)
			if err != nil {
				return microerror.Mask(err)
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

func (r *Resource) detachVolume(ctx context.Context, volumeID string, attachment VolumeAttachment) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detaching volume %s from instance %s", volumeID, attachment.InstanceID))

	_, err := r.clients.EC2.DetachVolume(&ec2.DetachVolumeInput{
		Device:     aws.String(attachment.Device),
		Force:      aws.Bool(true),
		InstanceId: aws.String(attachment.InstanceID),
		VolumeId:   aws.String(volumeID),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detached volume %s from instance %s", volumeID, attachment.InstanceID))

	return nil
}

func (r *Resource) deleteVolume(ctx context.Context, volumeID string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting ebs volume %s", volumeID))

	deleteOperation := func() error {
		_, err := r.clients.EC2.DeleteVolume(&ec2.DeleteVolumeInput{
			VolumeId: aws.String(volumeID),
		})
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	deleteNotify := func(err error, delay time.Duration) {
		r.logger.LogCtx(ctx, "level", "error", fmt.Sprintf("deleting ebs volume failed, retrying with delay %.0fm%.0fs: '%#v'", delay.Minutes(), delay.Seconds(), err))
	}
	deleteBackoff := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      3 * time.Minute,
		Clock:               backoff.SystemClock,
	}
	if err := backoff.RetryNotify(deleteOperation, deleteBackoff, deleteNotify); err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted ebs volume %s", volumeID))

	return nil
}
