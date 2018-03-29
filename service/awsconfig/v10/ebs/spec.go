package ebs

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

// Interface describes the methods provided by the helm client.
type Interface interface {
	// DeleteVolume deletes an EBS volume with retry logic.
	DeleteVolume(ctx context.Context, volumeID string) error
	// DetachVolume detaches an EBS volume. If force is specified data loss may
	// occur. If shutdown is specified the instance will be stopped first.
	DetachVolume(ctx context.Context, volumeID string, attachment VolumeAttachment, force bool, shutdown bool) error
	// ListVolumes lists EBS volumes for a guest cluster. If etcdVolume is true
	// the Etcd volume for the master instance will be returned. If
	// persistentVolume is true then any Persistent Volumes associated with the
	// cluster will be returned.
	ListVolumes(customObject v1alpha1.AWSConfig, etcdVolume bool, persistentVolume bool) ([]Volume, error)
}

// EC2Client describes the methods required to be implemented by an EC2 AWS client.
type EC2Client interface {
	DeleteVolume(*ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
	DescribeVolumes(*ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	DetachVolume(*ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error)
	StopInstances(*ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error)
}

// Volume is an EBS volume and its attachments.
type Volume struct {
	VolumeID    string
	Attachments []VolumeAttachment
}

// VolumeAttachment is an EBS volume attached to an EC2 instance.
type VolumeAttachment struct {
	Device     string
	InstanceID string
}
