package adapter

import (
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v10/templates/cloudformation/guest/outputs.go
//

type outputsAdapter struct {
	Master        outputsAdapterMaster
	Worker        outputsAdapterWorker
	VersionBundle outputsAdapterVersionBundle
}

type outputsAdapterMaster struct {
	ImageID     string
	Instance    outputsAdapterMasterInstance
	CloudConfig outputsAdapterMasterCloudConfig
}

type outputsAdapterMasterInstance struct {
	ResourceName string
	Type         string
}

type outputsAdapterMasterCloudConfig struct {
	Version string
}

type outputsAdapterWorker struct {
	ASG          outputsAdapterWorkerASG
	Count        string
	ImageID      string
	InstanceType string
	CloudConfig  outputsAdapterWorkerCloudConfig
}

type outputsAdapterWorkerASG struct {
	Key string
	Ref string
}

type outputsAdapterWorkerCloudConfig struct {
	Version string
}

type outputsAdapterVersionBundle struct {
	Version string
}

func (a *outputsAdapter) Adapt(config Config) error {
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
