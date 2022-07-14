package ebs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v2/pkg/awstags"
	"github.com/giantswarm/aws-operator/v2/service/controller/key"
)

const (
	cloudProviderClusterTagValue = "owned"
)

type Config struct {
	Client EC2Client
	Logger micrologger.Logger
}

type EBS struct {
	client EC2Client
	logger micrologger.Logger
}

func New(config Config) (*EBS, error) {
	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	e := &EBS{
		client: config.Client,
		logger: config.Logger,
	}

	return e, nil
}

// DeleteVolume deletes an EBS volume with retry logic.
func (e *EBS) DeleteVolume(ctx context.Context, volumeID string) error {
	e.logger.Debugf(ctx, "deleting EBS volume %#q", volumeID)

	o := func() error {
		i := &ec2.DeleteVolumeInput{
			VolumeId: aws.String(volumeID),
		}

		_, err := e.client.DeleteVolume(i)
		if IsVolumeNotFound(err) {
			// Fall through.
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(30*time.Second, 5*time.Second)
	n := backoff.NewNotifier(e.logger, context.Background())

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	e.logger.Debugf(ctx, "deleted EBS volume %#q", volumeID)

	return nil
}

// DetachVolume detaches an EBS volume. If force is specified data loss may
// occur. If shutdown is specified the instance will be stopped first. Note that
// the long term plan for the interface and functionality is to remove the EC2
// instance management so the EBS service implementation here only manages EBS
// volumes.
func (e *EBS) DetachVolume(ctx context.Context, volumeID string, attachment VolumeAttachment, force bool, shutdown bool, wait bool) error {
	if shutdown {
		e.logger.Debugf(ctx, "requesting to stop instance %#q", attachment.InstanceID)

		i := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(attachment.InstanceID),
			},
		}

		_, err := e.client.StopInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.Debugf(ctx, "requested to stop instance %#q", attachment.InstanceID)
	}

	if shutdown && wait {
		e.logger.Debugf(ctx, "waiting for instance %#q to be stopped", attachment.InstanceID)

		i := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(attachment.InstanceID),
			},
		}

		err := e.client.WaitUntilInstanceStopped(i)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.Debugf(ctx, "waited for instance %#q to be stopped", attachment.InstanceID)
	}

	{
		e.logger.Debugf(ctx, "requesting EBS volume %#q to be detached from instance %#q", volumeID, attachment.InstanceID)

		i := &ec2.DetachVolumeInput{
			Device:     aws.String(attachment.Device),
			InstanceId: aws.String(attachment.InstanceID),
			VolumeId:   aws.String(volumeID),
			Force:      aws.Bool(force),
		}

		_, err := e.client.DetachVolume(i)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.Debugf(ctx, "requested EBS volume %#q to be detached from instance %#q", volumeID, attachment.InstanceID)
	}

	{
		i := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{
				aws.String(volumeID),
			},
		}

		v, err := e.client.DescribeVolumes(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(v.Volumes) == 0 {
			// Fall through.
			return microerror.Maskf(executionFailedError, "no volume found")
		}

		if len(v.Volumes) > 1 {
			return microerror.Maskf(executionFailedError, "more than one volume has been found")
		}

		d := v.Volumes[0]
		if len(d.Attachments) > 1 {
			// Shouldn't happen; log so we know if it is
			e.logger.Debugf(ctx, "found multiple attachment for volume %#q", volumeID)
		}

		for _, a := range d.Attachments {
			if a.State != nil {
				attachmentStatus := *a.State
				if attachmentStatus != ec2.VolumeAttachmentStateDetached {
					return microerror.Maskf(volumeAttachedError, "volume %#q still attached", volumeID)
				}
			} else {
				// Shouldn't happen; log so we know if it is
				e.logger.Debugf(ctx, "nil attachment state for volume %#q", volumeID)
			}
		}

		e.logger.Debugf(ctx, "detached EBS volume %#q from instance %#q", volumeID, attachment.InstanceID)
	}

	return nil
}

// ListVolumes lists EBS volumes for a guest cluster. If etcdVolume is true
// the Etcd volume for the master instance will be returned. If persistentVolume
// is set then any Persistent Volumes associated with the cluster will be
// returned.
func (e *EBS) ListVolumes(ctx context.Context, cr infrastructurev1alpha3.AWSCluster, filterFuncs ...func(t *ec2.Tag) bool) ([]Volume, error) {
	var volumes []Volume

	// We filter to only select clusters with the cluster cloud provider tag.
	i := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.ClusterCloudProviderTag(&cr))),
				Values: []*string{
					aws.String(cloudProviderClusterTagValue),
				},
			},
		},
	}

	o, err := e.client.DescribeVolumes(i)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, v := range o.Volumes {
		if !IsFiltered(v, filterFuncs) {
			continue
		}
		ignoreVolume := false
		attachments := []VolumeAttachment{}

		if len(v.Attachments) > 0 {
			for _, a := range v.Attachments {

				if *a.InstanceId != "" {
					i := &ec2.DescribeInstancesInput{
						InstanceIds: aws.StringSlice([]string{*a.InstanceId}),
					}

					o, err := e.client.DescribeInstances(i)
					if err != nil {
						return nil, microerror.Mask(err)
					}

					hasWrongClusterID := false
					if len(o.Reservations) > 0 && len(o.Reservations[0].Instances) > 0 {
						hasWrongClusterID = awstags.ValueForKey(o.Reservations[0].Instances[0].Tags, key.TagCluster) != cr.Labels[key.TagCluster]
					}

					// if the instance does not have the proper cluster-id tag than ignore the volume
					if hasWrongClusterID {
						e.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("EBS volume %#q is attached to instance %#q which does not belong to the cluster %#q, ignoring the volume", *a.VolumeId, *a.InstanceId, cr.Labels[key.TagCluster]))

						// do not add this volume to the volume list
						// this volume will be ignored by the operator
						ignoreVolume = true
						break
					}
				}

				attachments = append(attachments, VolumeAttachment{
					Device:     *a.Device,
					InstanceID: *a.InstanceId,
				})
			}
		}

		if ignoreVolume {
			continue
		}

		volume := Volume{
			VolumeID:    *v.VolumeId,
			Attachments: attachments,
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}
