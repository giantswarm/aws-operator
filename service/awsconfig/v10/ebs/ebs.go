package ebs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

const (
	cloudProviderClusterTagValue        = "owned"
	cloudProviderPersistentVolumeTagKey = "kubernetes.io/created-for/pv/name"
	nameTagKey                          = "Name"
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
	e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting EBS volume %s", volumeID))

	deleteOperation := func() error {
		_, err := e.client.DeleteVolume(&ec2.DeleteVolumeInput{
			VolumeId: aws.String(volumeID),
		})
		if IsVolumeNotFound(err) {
			// Fall through.
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	deleteNotify := func(err error, delay time.Duration) {
		e.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("deleting EBS volume failed, retrying with delay %s", delay.String()))
	}
	deleteBackoff := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      30 * time.Second,
		Clock:               backoff.SystemClock,
	}
	if err := backoff.RetryNotify(deleteOperation, deleteBackoff, deleteNotify); err != nil {
		return microerror.Mask(err)
	}

	e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted EBS volume %s", volumeID))

	return nil
}

// DetachVolume detaches an EBS volume. If force is specified data loss may occur. If shutdown is
// specified the instance will be stopped first.
func (e *EBS) DetachVolume(ctx context.Context, volumeID string, attachment VolumeAttachment, force bool, shutdown bool) error {
	e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detaching EBS volume %s from instance %s", volumeID, attachment.InstanceID))

	if shutdown {
		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("stopping instance %s", attachment.InstanceID))

		_, err := e.client.StopInstances(&ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(attachment.InstanceID),
			},
		})
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("stopped instance %s", attachment.InstanceID))
	}

	_, err := e.client.DetachVolume(&ec2.DetachVolumeInput{
		Device:     aws.String(attachment.Device),
		InstanceId: aws.String(attachment.InstanceID),
		VolumeId:   aws.String(volumeID),
		Force:      aws.Bool(force),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detached EBS volume %s from instance %s", volumeID, attachment.InstanceID))

	return nil
}

// ListVolumes lists EBS volumes for a guest cluster. If etcdVolume is true
// the Etcd volume for the master instance will be returned. If persistentVolume
// is set then any Persistent Volumes associated with the cluster will be
// returned.
func (e *EBS) ListVolumes(customObject v1alpha1.AWSConfig, etcdVolume bool, persistentVolume bool) ([]Volume, error) {
	etcdVolumeName := ""
	volumes := []Volume{}

	// We filter to only select clusters with the cluster cloud provider tag.
	clusterTag := key.ClusterCloudProviderTag(customObject)
	filters := []*ec2.Filter{
		{
			Name: aws.String(fmt.Sprintf("tag:%s", clusterTag)),
			Values: []*string{
				aws.String(cloudProviderClusterTagValue),
			},
		},
	}
	output, err := e.client.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Set the volume name so we select the correct volume.
	if etcdVolume {
		etcdVolumeName = key.EtcdVolumeName(customObject)
	}

	for _, vol := range output.Volumes {
		// Volume is only returned if it has the correct tags.
		if containsVolumeTag(vol.Tags, etcdVolumeName, persistentVolume) {
			attachments := []VolumeAttachment{}

			if len(vol.Attachments) > 0 {
				for _, a := range vol.Attachments {
					attachments = append(attachments, VolumeAttachment{
						Device:     *a.Device,
						InstanceID: *a.InstanceId,
					})
				}
			}

			volume := Volume{
				VolumeID:    *vol.VolumeId,
				Attachments: attachments,
			}

			volumes = append(volumes, volume)
		}
	}

	return volumes, nil
}

func containsVolumeTag(tags []*ec2.Tag, etcdVolumeName string, persistentVolume bool) bool {
	for _, tag := range tags {
		if etcdVolumeName != "" && *tag.Key == nameTagKey && *tag.Value == etcdVolumeName {
			return true
		}

		if persistentVolume && *tag.Key == cloudProviderPersistentVolumeTagKey {
			return true
		}
	}

	return false
}
