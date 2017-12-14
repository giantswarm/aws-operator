package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/launch_configuration.go

type launchConfigAdapter struct {
	AssociatePublicIPAddress bool
	BlockDeviceMappings      []BlockDeviceMapping
	IAMInstanceProfileName   string
	ImageID                  string
	InstanceType             string
	SecurityGroupID          string
	SmallCloudConfig         string
}

type BlockDeviceMapping struct {
	DeleteOnTermination bool
	DeviceName          string
	VolumeSize          int
	VolumeType          string
}

func (l *launchConfigAdapter) getLaunchConfiguration(customObject v1alpha1.AWSConfig, clients Clients) error {
	if len(customObject.Spec.AWS.Workers) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	l.ImageID = keyv2.WorkerImageID(customObject)
	l.InstanceType = keyv2.WorkerInstanceType(customObject)
	l.IAMInstanceProfileName = keyv2.InstanceProfileName(customObject, prefixWorker)
	l.AssociatePublicIPAddress = false

	l.BlockDeviceMappings = []BlockDeviceMapping{
		BlockDeviceMapping{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          defaultEBSVolumeSize,
			VolumeType:          defaultEBSVolumeType,
		},
	}

	// security group
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
	groupName := keyv2.SecurityGroupName(customObject, prefixWorker)
	describeSgInput := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(groupName),
				},
			},
			{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(groupName),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeSecurityGroups(describeSgInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.SecurityGroups) > 1 {
		return microerror.Mask(tooManyResultsError)
	}
	l.SecurityGroupID = *output.SecurityGroups[0].GroupId

	// cloud config
	accountID, err := AccountID(clients)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := keyv2.ClusterID(customObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := SmallCloudconfigConfig{
		MachineType:    prefixWorker,
		Region:         customObject.Spec.AWS.Region,
		S3URI:          s3URI,
		ClusterVersion: keyv2.ClusterVersion(customObject),
	}
	smallCloudConfig, err := SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	l.SmallCloudConfig = smallCloudConfig

	return nil
}
