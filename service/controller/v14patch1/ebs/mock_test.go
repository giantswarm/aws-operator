package ebs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v14/key"
)

type EC2ClientMock struct {
	customObject v1alpha1.AWSConfig
	ebsVolumes   []ebsVolumeMock
}

type ebsVolumeMock struct {
	volumeID    string
	attachments []*ec2.VolumeAttachment
	tags        []*ec2.Tag
}

func (e *EC2ClientMock) DeleteVolume(*ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
	return nil, nil
}

func (e *EC2ClientMock) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	output := &ec2.DescribeVolumesOutput{}
	volumes := []*ec2.Volume{}

	clusterTag := key.ClusterCloudProviderTag(e.customObject)

	for _, mock := range e.ebsVolumes {
		vol := &ec2.Volume{
			VolumeId:    aws.String(mock.volumeID),
			Attachments: mock.attachments,
			Tags:        mock.tags,
		}

		for _, tag := range mock.tags {
			if *tag.Key == clusterTag && *tag.Value == "owned" {
				volumes = append(volumes, vol)
			}
		}
	}

	output.SetVolumes(volumes)

	return output, nil
}

func (e *EC2ClientMock) DetachVolume(*ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
	return nil, nil
}

func (e *EC2ClientMock) StopInstances(*ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error) {
	return nil, nil
}

func (e *EC2ClientMock) WaitUntilInstanceStopped(*ec2.DescribeInstancesInput) error {
	return nil
}
