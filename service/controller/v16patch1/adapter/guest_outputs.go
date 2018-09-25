package adapter

import (
	"strconv"

	"github.com/giantswarm/aws-operator/service/controller/v16patch1/key"
)

type GuestOutputsAdapter struct {
	Master         GuestOutputsAdapterMaster
	Worker         GuestOutputsAdapterWorker
	Route53Enabled bool
	VersionBundle  GuestOutputsAdapterVersionBundle
}

func (a *GuestOutputsAdapter) Adapt(config Config) error {
	a.Route53Enabled = route53Enabled(config)
	a.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
	a.Master.ImageID = config.StackState.MasterImageID
	a.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
	a.Master.Instance.Type = config.StackState.MasterInstanceType
	a.Master.CloudConfig.Version = config.StackState.MasterCloudConfigVersion

	a.Worker.ASG.Key = key.WorkerASGKey
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.Count = config.StackState.WorkerCount
	a.Worker.DockerVolumeSizeGB = strconv.Itoa(config.StackState.WorkerDockerVolumeSizeGB)
	a.Worker.ImageID = config.StackState.WorkerImageID
	a.Worker.InstanceType = config.StackState.WorkerInstanceType
	a.Worker.CloudConfig.Version = config.StackState.WorkerCloudConfigVersion

	a.VersionBundle.Version = config.StackState.VersionBundleVersion

	return nil
}

type GuestOutputsAdapterMaster struct {
	ImageID      string
	Instance     GuestOutputsAdapterMasterInstance
	CloudConfig  GuestOutputsAdapterMasterCloudConfig
	DockerVolume GuestOutputsAdapterMasterDockerVolume
}

type GuestOutputsAdapterMasterInstance struct {
	ResourceName string
	Type         string
}

type GuestOutputsAdapterMasterCloudConfig struct {
	Version string
}

type GuestOutputsAdapterMasterDockerVolume struct {
	ResourceName string
}

type GuestOutputsAdapterWorker struct {
	ASG                GuestOutputsAdapterWorkerASG
	Count              string
	DockerVolumeSizeGB string
	ImageID            string
	InstanceType       string
	CloudConfig        GuestOutputsAdapterWorkerCloudConfig
}

type GuestOutputsAdapterWorkerASG struct {
	Key string
	Ref string
}

type GuestOutputsAdapterWorkerCloudConfig struct {
	Version string
}

type GuestOutputsAdapterVersionBundle struct {
	Version string
}
