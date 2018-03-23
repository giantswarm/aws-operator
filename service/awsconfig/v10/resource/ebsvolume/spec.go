package ebsvolume

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Clients struct {
	EC2 EC2Client
}

// EC2Client describes the methods required to be implemented by an EC2 AWS client.
type EC2Client interface {
	DeleteVolume(*ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
	DescribeVolumes(*ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	DetachVolume(*ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error)
}

type EBSVolumeState struct {
	Volumes []Volume
}

type Volume struct {
	VolumeID    string
	Attachments []VolumeAttachment
}

type VolumeAttachment struct {
	Device     string
	InstanceID string
}
