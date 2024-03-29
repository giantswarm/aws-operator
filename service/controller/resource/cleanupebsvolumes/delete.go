package cleanupebsvolumes

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v16/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v16/service/controller/key"
	"github.com/giantswarm/aws-operator/v16/service/internal/ebs"
)

// EnsureDeleted detaches and deletes the EBS volumes. We don't return
// errors so deletion logic in following resources is executed.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var ebsService ebs.Interface
	{
		c := ebs.Config{
			Client: cc.Client.TenantCluster.AWS.EC2,
			Logger: r.logger,
		}

		ebsService, err = ebs.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Get all etcd, docker and persistent volumes.
	filterFuncs := []func(t *ec2.Tag) bool{
		ebs.NewDockerVolumeFilter(cr),
		ebs.NewEtcdVolumeFilter(cr),
		ebs.NewPersistentVolumeFilter(cr),
	}
	volumes, err := ebsService.ListVolumes(ctx, cr, filterFuncs...)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(volumes) > 0 {
		r.logger.Debugf(ctx, "deleting %d EBS volumes", len(volumes))

		// First detach any attached volumes without forcing but shutdown the
		// instances.
		for _, vol := range volumes {
			for _, a := range vol.Attachments {
				force := false
				shutdown := true
				wait := false
				err := ebsService.DetachVolume(ctx, vol.VolumeID, a, force, shutdown, wait)
				if ebs.IsVolumeAttached(err) {
					r.logger.Debugf(ctx, "volume is still attached")
					r.logger.Debugf(ctx, "canceling resource")
					return nil
				} else if err != nil {
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
				if ebs.IsVolumeAttached(err) {
					r.logger.Debugf(ctx, "volume is still attached")
					r.logger.Debugf(ctx, "canceling resource")
					return nil
				} else if err != nil {
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

		r.logger.Debugf(ctx, "deleted %d EBS volumes", len(volumes))
	} else {
		r.logger.Debugf(ctx, "not deleting EBS volumes because there aren't any")
	}

	return nil
}
