package adapter

import "github.com/giantswarm/aws-operator/service/cloudconfigv3"

// template related to this adapter: service/templates/cloudformation/guest/outputs.yaml

type outputsAdapter struct {
	MasterCloudConfigVersion string
	WorkerCloudConfigVersion string
}

func (o *outputsAdapter) getOutputs(cfg Config) error {
	o.MasterCloudConfigVersion = cloudconfigv3.MasterCloudConfigVersion
	o.WorkerCloudConfigVersion = cloudconfigv3.WorkerCloudConfigVersion

	return nil
}
