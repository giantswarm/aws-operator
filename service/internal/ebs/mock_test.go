package ebs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type EC2ClientMock struct {
	customObject infrastructurev1alpha3.AWSCluster
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

func (e *EC2ClientMock) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	o := &ec2.DescribeInstancesOutput{}

	// test case for instance that does not belong to the cluster
	// to test behavior of ignoring volume when its mounted to an instance from different cluster
	if *input.InstanceIds[0] == "i-555555" {

		t := &ec2.Tag{
			Key:   aws.String(key.TagCluster),
			Value: aws.String("invalid-cluster"),
		}
		i := &ec2.Instance{
			Tags: []*ec2.Tag{t},
		}
		r := &ec2.Reservation{
			Instances: []*ec2.Instance{i},
		}

		o.Reservations = []*ec2.Reservation{r}
	}

	return o, nil
}

func (e *EC2ClientMock) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	output := &ec2.DescribeVolumesOutput{}
	volumes := []*ec2.Volume{}

	clusterTag := key.ClusterCloudProviderTag(&e.customObject)

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
