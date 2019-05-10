package adapter

import (
	"encoding/base64"
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates"
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
	VolumeSize          string
	VolumeType          string
}

func (l *GuestLaunchConfigAdapter) Adapt(config Config) error {
	l.ASGType = "worker"
	l.WorkerInstanceType = key.WorkerInstanceType(config.MachineDeployment)
	l.WorkerImageID = config.StackState.WorkerImageID
	l.WorkerAssociatePublicIPAddress = false

	{
		cur := config.StackState.WorkerDockerVolumeSizeGB
		def := defaultEBSVolumeSize

		if cur == "" {
			cur = "0"
		}

		curi, err := strconv.Atoi(cur)
		if err != nil {
			return microerror.Mask(err)
		}

		if curi <= 0 {
			config.StackState.WorkerDockerVolumeSizeGB = def
		}
	}

	{
		cur := config.StackState.WorkerLogVolumeSizeGB
		def := defaultEBSVolumeSize

		if cur == "" {
			cur = "0"
		}

		curi, err := strconv.Atoi(cur)
		if err != nil {
			return microerror.Mask(err)
		}

		if curi <= 0 {
			config.StackState.WorkerLogVolumeSizeGB = def
		}
	}

	{
		cur := config.StackState.WorkerKubeletVolumeSizeGB
		def := defaultEBSVolumeSize

		if cur == "" {
			cur = "0"
		}

		curi, err := strconv.Atoi(cur)
		if err != nil {
			return microerror.Mask(err)
		}

		if curi <= 0 {
			config.StackState.WorkerKubeletVolumeSizeGB = def
		}
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
		{
			// TL;DR; kubelet volume same size as docker volume
			// this is a temporary solution that should stay around until Node Pools story is ready.
			// See here for furhter info https://github.com/giantswarm/giantswarm/issues/5582#issuecomment-476170597
			DeleteOnTermination: true,
			DeviceName:          kubeletEBSVolumeMountPoint,
			VolumeSize:          config.StackState.WorkerKubeletVolumeSizeGB,
			VolumeType:          defaultEBSVolumeType,
		},
	}
	l.WorkerInstanceMonitoring = config.StackState.WorkerInstanceMonitoring

	// small cloud config field.
	c := SmallCloudconfigConfig{
		InstanceRole: "worker",
		S3URL:        key.SmallCloudConfigS3URL(config.CustomObject, config.TenantClusterAccountID, "worker"),
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	l.WorkerSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
