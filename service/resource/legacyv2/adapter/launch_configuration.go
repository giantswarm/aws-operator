package adapter

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
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

func (l *launchConfigAdapter) getLaunchConfiguration(customObject v1alpha1.AWSConfig, clients Clients) error {
	if len(customObject.Spec.AWS.Workers) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	l.WorkerImageID = keyv2.WorkerImageID(customObject)
	l.WorkerInstanceType = keyv2.WorkerInstanceType(customObject)
	l.WorkerAssociatePublicIPAddress = false

	l.WorkerBlockDeviceMappings = []BlockDeviceMapping{
		BlockDeviceMapping{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          defaultEBSVolumeSize,
			VolumeType:          defaultEBSVolumeType,
		},
	}

	// small cloud config field.
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
	l.WorkerSmallCloudConfig = smallCloudConfig

	return nil
}
