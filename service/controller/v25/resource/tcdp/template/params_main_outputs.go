package template

import (
	"strconv"

	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

type ParamsMainOutputs struct {
	Master         ParamsMainOutputsMaster
	Worker         ParamsMainOutputsWorker
	Route53Enabled bool
	VersionBundle  ParamsMainOutputsVersionBundle
}

func (a *ParamsMainOutputs) Adapt(config Config) error {
	a.Route53Enabled = config.Route53Enabled
	a.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
	a.Master.ImageID = config.StackState.MasterImageID
	a.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
	a.Master.Instance.Type = config.StackState.MasterInstanceType
	a.Master.CloudConfig.Version = config.StackState.MasterCloudConfigVersion

	a.Worker.ASG.Key = key.WorkerASGKey
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.CloudConfig.Version = config.StackState.WorkerCloudConfigVersion
	a.Worker.DockerVolumeSizeGB = strconv.Itoa(config.StackState.WorkerDockerVolumeSizeGB)
	a.Worker.ImageID = config.StackState.WorkerImageID
	a.Worker.InstanceType = config.StackState.WorkerInstanceType

	a.VersionBundle.Version = config.StackState.VersionBundleVersion

	return nil
}

type ParamsMainOutputsMaster struct {
	ImageID      string
	Instance     ParamsMainOutputsMasterInstance
	CloudConfig  ParamsMainOutputsMasterCloudConfig
	DockerVolume ParamsMainOutputsMasterDockerVolume
}

type ParamsMainOutputsMasterInstance struct {
	ResourceName string
	Type         string
}

type ParamsMainOutputsMasterCloudConfig struct {
	Version string
}

type ParamsMainOutputsMasterDockerVolume struct {
	ResourceName string
}

type ParamsMainOutputsWorker struct {
	ASG                ParamsMainOutputsWorkerASG
	CloudConfig        ParamsMainOutputsWorkerCloudConfig
	DockerVolumeSizeGB string
	ImageID            string
	InstanceType       string
}

type ParamsMainOutputsWorkerASG struct {
	Key string
	Ref string
}

type ParamsMainOutputsWorkerCloudConfig struct {
	Version string
}

type ParamsMainOutputsVersionBundle struct {
	Version string
}
