package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/launch_configuration.go

type launchConfigAdapter struct {
	WorkerAssociatePublicIPAddress bool
	WorkerBlockDeviceMappings      []BlockDeviceMapping
	WorkerImageID                  string
	WorkerInstanceType             string
	WorkerSecurityGroupID          string
	WorkerSmallCloudConfig         string
}

type BlockDeviceMapping struct {
	DeleteOnTermination bool
	DeviceName          string
	VolumeSize          int
	VolumeType          string
}

func (l *launchConfigAdapter) getLaunchConfiguration(cfg Config) error {
	if len(cfg.CustomObject.Spec.AWS.Workers) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	l.WorkerImageID = keyv2.WorkerImageID(cfg.CustomObject)
	l.WorkerInstanceType = keyv2.WorkerInstanceType(cfg.CustomObject)
	l.WorkerAssociatePublicIPAddress = false

	l.WorkerBlockDeviceMappings = []BlockDeviceMapping{
		BlockDeviceMapping{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          defaultEBSVolumeSize,
			VolumeType:          defaultEBSVolumeType,
		},
	}

	// security group field.
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
	groupName := keyv2.SecurityGroupName(cfg.CustomObject, prefixWorker)
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
	output, err := cfg.Clients.EC2.DescribeSecurityGroups(describeSgInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.SecurityGroups) > 1 {
		return microerror.Mask(tooManyResultsError)
	}
	l.WorkerSecurityGroupID = *output.SecurityGroups[0].GroupId

	// small cloud config field.
	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	clusterID := keyv2.ClusterID(cfg.CustomObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := SmallCloudconfigConfig{
		MachineType:    prefixWorker,
		Region:         cfg.CustomObject.Spec.AWS.Region,
		S3URI:          s3URI,
		ClusterVersion: keyv2.ClusterVersion(cfg.CustomObject),
	}
	smallCloudConfig, err := SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = smallCloudConfig

	return nil
}
