package adapter

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v7/cloudconfig"
	"github.com/giantswarm/aws-operator/service/awsconfig/v7/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v7/templates/cloudformation/guest/outputs.go
//

type outputsAdapter struct {
	Master        outputsAdapterMaster
	Worker        outputsAdapterWorker
	VersionBundle outputsAdapterVersionBundle
}

type outputsAdapterMaster struct {
	ImageID      string
	InstanceType string
	CloudConfig  outputsAdapterMasterCloudConfig
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
	Tag string
}

type outputsAdapterWorkerCloudConfig struct {
	Version string
}

type outputsAdapterVersionBundle struct {
	Version string
}

func (a *outputsAdapter) Adapt(config Config) error {
	imageID, err := key.ImageID(config.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}
	workerCount := key.WorkerCount(config.CustomObject)
	if workerCount <= 0 {
		return microerror.Maskf(invalidConfigError, "at least 1 worker required, found %d", workerCount)
	}

	a.Master.ImageID = imageID
	a.Master.InstanceType = key.MasterInstanceType(config.CustomObject)
	a.Master.CloudConfig.Version = cloudconfig.MasterCloudConfigVersion

	a.Worker.ASG.Tag = key.WorkerASGTag
	a.Worker.Count = strconv.Itoa(workerCount)
	a.Worker.ImageID = imageID
	a.Worker.InstanceType = key.WorkerInstanceType(config.CustomObject)
	a.Worker.CloudConfig.Version = cloudconfig.WorkerCloudConfigVersion

	a.VersionBundle.Version = key.VersionBundleVersion(config.CustomObject)

	return nil
}
