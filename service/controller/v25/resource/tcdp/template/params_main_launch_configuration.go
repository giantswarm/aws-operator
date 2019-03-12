package template

import (
	"encoding/base64"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v25/key"
	"github.com/giantswarm/aws-operator/service/controller/v25/templates"
)

type ParamsMainLaunchConfig struct {
	ASGType                        string
	WorkerAssociatePublicIPAddress bool
	WorkerBlockDeviceMappings      []BlockDeviceMapping
	WorkerInstanceMonitoring       bool
	WorkerInstanceType             string
	WorkerImageID                  string
	WorkerSecurityGroupID          string
	WorkerSmallCloudConfig         string
}

type BlockDeviceMapping struct {
	DeleteOnTermination bool
	DeviceName          string
	VolumeSize          int
	VolumeType          string
}

func (l *ParamsMainLaunchConfig) Adapt(config Config) error {
	l.ASGType = key.KindWorker
	l.WorkerInstanceType = key.WorkerInstanceType(config.CustomObject)
	l.WorkerImageID = config.StackState.WorkerImageID
	l.WorkerAssociatePublicIPAddress = false

	if config.StackState.WorkerDockerVolumeSizeGB <= 0 {
		config.StackState.WorkerDockerVolumeSizeGB = defaultEBSVolumeSize
	}

	if config.StackState.WorkerLogVolumeSizeGB <= 0 {
		config.StackState.WorkerLogVolumeSizeGB = defaultEBSVolumeSize
	}

	l.WorkerBlockDeviceMappings = []BlockDeviceMapping{
		{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          config.StackState.WorkerDockerVolumeSizeGB,
			VolumeType:          defaultEBSVolumeType,
		},
		{
			DeleteOnTermination: true,
			DeviceName:          logEBSVolumeMountPoint,
			VolumeSize:          config.StackState.WorkerLogVolumeSizeGB,
			VolumeType:          defaultEBSVolumeType,
		},
	}
	l.WorkerInstanceMonitoring = config.StackState.WorkerInstanceMonitoring

	// small cloud config field.
	c := SmallCloudconfigConfig{
		InstanceRole: key.KindWorker,
		S3URL:        key.SmallCloudConfigS3URL(config.CustomObject, config.TenantClusterAccountID, key.KindWorker),
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
