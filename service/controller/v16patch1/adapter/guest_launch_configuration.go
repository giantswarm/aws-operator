package adapter

import (
	"encoding/base64"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v16patch1/key"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1/templates"
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
	l.ASGType = asgType(config)
	l.WorkerInstanceType = key.WorkerInstanceType(config.CustomObject)
	l.WorkerImageID = workerImageID(config)
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
		Region:    key.Region(config.CustomObject),
		Registry:  key.AWSCliContainerRegistry(config.CustomObject),
		Role:      prefixWorker,
		S3HTTPURL: key.SmallCloudConfigS3HTTPURL(config.CustomObject, accountID, prefixWorker),
		S3URL:     key.SmallCloudConfigS3URL(config.CustomObject, accountID, prefixWorker),
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
