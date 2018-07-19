package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/v14/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v14/templates/cloudformation/guest/outputs.go
//

type guestOutputsAdapter struct {
	Master         guestOutputsAdapterMaster
	Worker         guestOutputsAdapterWorker
	Route53Enabled bool
	VersionBundle  guestOutputsAdapterVersionBundle
}

func (a *guestOutputsAdapter) Adapt(config Config) error {
	a.Route53Enabled = route53Enabled(config)
	a.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
	a.Master.ImageID = config.StackState.MasterImageID
	a.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
	a.Master.Instance.Type = config.StackState.MasterInstanceType
	a.Master.CloudConfig.Version = config.StackState.MasterCloudConfigVersion

	a.Worker.ASG.Key = key.WorkerASGKey
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.Count = config.StackState.WorkerCount
	a.Worker.ImageID = config.StackState.WorkerImageID
	a.Worker.InstanceType = config.StackState.WorkerInstanceType
	a.Worker.CloudConfig.Version = config.StackState.WorkerCloudConfigVersion

	a.VersionBundle.Version = config.StackState.VersionBundleVersion

	return nil
}

type guestOutputsAdapterMaster struct {
	ImageID      string
	Instance     guestOutputsAdapterMasterInstance
	CloudConfig  guestOutputsAdapterMasterCloudConfig
	DockerVolume guestOutputsAdapterMasterDockerVolume
}

type guestOutputsAdapterMasterInstance struct {
	ResourceName string
	Type         string
}

type guestOutputsAdapterMasterCloudConfig struct {
	Version string
}

type guestOutputsAdapterMasterDockerVolume struct {
	ResourceName string
}

type guestOutputsAdapterWorker struct {
	ASG          guestOutputsAdapterWorkerASG
	Count        string
	ImageID      string
	InstanceType string
	CloudConfig  guestOutputsAdapterWorkerCloudConfig
}

type guestOutputsAdapterWorkerASG struct {
	Key string
	Ref string
}

type guestOutputsAdapterWorkerCloudConfig struct {
	Version string
}

type guestOutputsAdapterVersionBundle struct {
	Version string
}
