package ebs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting EBS volume %#q", volumeID))

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

	e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted EBS volume %#q", volumeID))

	return nil
}

// DetachVolume detaches an EBS volume. If force is specified data loss may occur. If shutdown is
// specified the instance will be stopped first.
func (e *EBS) DetachVolume(ctx context.Context, volumeID string, attachment VolumeAttachment, force bool, shutdown bool, wait bool) error {
	if shutdown {
		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("requesting to stop instance %#q", attachment.InstanceID))

		i := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(attachment.InstanceID),
			},
		}

		_, err := e.client.StopInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("requested to stop instance %#q", attachment.InstanceID))
	}

	if shutdown && wait {
		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for instance %#q to stop", attachment.InstanceID))

		i := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(attachment.InstanceID),
			},
		}

		err := e.client.WaitUntilInstanceStopped(i)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for instance %#q to stop", attachment.InstanceID))
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detaching EBS volume %#q from instance %#q", volumeID, attachment.InstanceID))

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

		e.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detached EBS volume %#q from instance %#q", volumeID, attachment.InstanceID))
	}

	return nil
}

// ListVolumes lists EBS volumes for a guest cluster. If etcdVolume is true
// the Etcd volume for the master instance will be returned. If persistentVolume
// is set then any Persistent Volumes associated with the cluster will be
// returned.
func (e *EBS) ListVolumes(cr v1alpha1.Cluster, filterFuncs ...func(t *ec2.Tag) bool) ([]Volume, error) {
	var volumes []Volume

	// We filter to only select clusters with the cluster cloud provider tag.
	i := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.ClusterCloudProviderTag(cr))),
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

		attachments := []VolumeAttachment{}

		if len(v.Attachments) > 0 {
			for _, a := range v.Attachments {
				attachments = append(attachments, VolumeAttachment{
					Device:     *a.Device,
					InstanceID: *a.InstanceId,
				})
			}
		}

		volume := Volume{
			VolumeID:    *v.VolumeId,
			Attachments: attachments,
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}
