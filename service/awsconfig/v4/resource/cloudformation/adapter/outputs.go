package adapter

import "github.com/giantswarm/aws-operator/service/awsconfig/v4/cloudconfig"

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v4/templates/cloudformation/guest/outputs.go
//

type outputsAdapter struct {
	MasterCloudConfigVersion string
	WorkerCloudConfigVersion string
}

func (o *outputsAdapter) getOutputs(cfg Config) error {
	o.MasterCloudConfigVersion = cloudconfig.MasterCloudConfigVersion
	o.WorkerCloudConfigVersion = cloudconfig.WorkerCloudConfigVersion

	return nil
}
