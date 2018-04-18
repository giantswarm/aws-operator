package adapter

import (
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v9/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v9/key"
	"github.com/giantswarm/aws-operator/service/controller/v9/templates"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v9/templates/cloudformation/guest/launch_configuration.go
//

type launchConfigAdapter struct {
	WorkerAssociatePublicIPAddress bool
	WorkerBlockDeviceMappings      []BlockDeviceMapping
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
	l.WorkerInstanceType = key.WorkerInstanceType(cfg.CustomObject)
	l.WorkerAssociatePublicIPAddress = false

	l.WorkerBlockDeviceMappings = []BlockDeviceMapping{
		{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          defaultEBSVolumeSize,
			VolumeType:          defaultEBSVolumeType,
		},
	}

	// small cloud config field.
	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	clusterID := key.ClusterID(cfg.CustomObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	c := SmallCloudconfigConfig{
		MachineType:        prefixWorker,
		Region:             cfg.CustomObject.Spec.AWS.Region,
		S3URI:              s3URI,
		CloudConfigVersion: cloudconfig.WorkerCloudConfigVersion,
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
