package adapter

import (
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch2/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v14patch2/key"
	"github.com/giantswarm/aws-operator/service/controller/v14patch2/templates"
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

func (l *GuestLaunchConfigAdapter) Adapt(cfg Config) error {
	l.ASGType = asgType(cfg)
	l.WorkerInstanceType = key.WorkerInstanceType(cfg.CustomObject)
	l.WorkerImageID = workerImageID(cfg)
	l.WorkerAssociatePublicIPAddress = false

	l.WorkerBlockDeviceMappings = []BlockDeviceMapping{
		{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          defaultEBSVolumeSize,
			VolumeType:          defaultEBSVolumeType,
		},
	}
	l.WorkerInstanceMonitoring = cfg.StackState.WorkerInstanceMonitoring

	// small cloud config field.
	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	clusterID := key.ClusterID(cfg.CustomObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	c := SmallCloudconfigConfig{
		MachineType:             prefixWorker,
		Region:                  key.Region(cfg.CustomObject),
		S3Domain:                key.S3ServiceDomain(cfg.CustomObject),
		S3URI:                   s3URI,
		CloudConfigVersion:      cloudconfig.CloudConfigVersion,
		AWSCliContainerRegistry: key.AWSCliContainerRegistry(cfg.CustomObject),
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
