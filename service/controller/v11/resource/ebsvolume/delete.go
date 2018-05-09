package ebsvolume

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	ebsservicecontext "github.com/giantswarm/aws-operator/service/controller/v11/context/ebsservice"
	"github.com/giantswarm/aws-operator/service/controller/v11/key"
)

// EnsureDeleted detaches and deletes the EBS volumes. We don't return
// errors so deletion logic in following resources is executed.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	ebsService, err := ebsservicecontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get both the Etcd volume and any Persistent Volumes.
	etcdVolume := true
	persistentVolume := true
	volumes, err := ebsService.ListVolumes(customObject, etcdVolume, persistentVolume)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(volumes) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %d EBS volumes", len(volumes)))

		// First detach any attached volumes without forcing but shutdown the
		// instances.
		for _, vol := range volumes {
			for _, a := range vol.Attachments {
				force := false
				shutdown := true
				wait := false
				err := ebsService.DetachVolume(ctx, vol.VolumeID, a, force, shutdown, wait)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("failed to detach EBS volume %s", vol.VolumeID), "stack", fmt.Sprintf("%#v", err))
				}
			}
		}

		// Now force detach so the volumes can be deleted cleanly. Instances
		// are already shutdown.
		for _, vol := range volumes {
			for _, a := range vol.Attachments {
				force := true
				shutdown := false
				wait := false
				err := ebsService.DetachVolume(ctx, vol.VolumeID, a, force, shutdown, wait)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("failed to force detach EBS volume %s", vol.VolumeID), "stack", fmt.Sprintf("%#v", err))
				}
			}
		}

		// Now delete the volumes.
		for _, vol := range volumes {
			err := ebsService.DeleteVolume(ctx, vol.VolumeID)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("failed to delete EBS volume %s", vol.VolumeID), "stack", fmt.Sprintf("%#v", err))
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted %d EBS volumes", len(volumes)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting EBS volumes because there aren't any")
	}

	return nil
}
