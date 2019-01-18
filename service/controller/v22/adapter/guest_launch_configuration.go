package adapter

import (
	"encoding/base64"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v22/key"
	"github.com/giantswarm/aws-operator/service/controller/v22/templates"
)

type GuestLaunchConfigAdapter struct {
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

func (l *GuestLaunchConfigAdapter) Adapt(config Config) error {
	l.ASGType = key.KindWorker
	l.WorkerInstanceType = key.WorkerInstanceType(config.CustomObject)
	l.WorkerImageID = config.StackState.WorkerImageID
	l.WorkerAssociatePublicIPAddress = false

	if config.StackState.WorkerDockerVolumeSizeGB <= 0 {
		config.StackState.WorkerDockerVolumeSizeGB = defaultEBSVolumeSize
	}

	l.WorkerBlockDeviceMappings = []BlockDeviceMapping{
		{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          config.StackState.WorkerDockerVolumeSizeGB,
			VolumeType:          defaultEBSVolumeType,
		},
	}
	l.WorkerInstanceMonitoring = config.StackState.WorkerInstanceMonitoring

	// small cloud config field.
	accountID, err := AccountID(config.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	c := SmallCloudconfigConfig{
		S3URL: key.SmallCloudConfigS3URL(config.CustomObject, accountID, key.KindWorker),
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
