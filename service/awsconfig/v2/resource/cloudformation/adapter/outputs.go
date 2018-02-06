package adapter

// NOTE(PK): This import is disturbing. I'm not bothering. It's first candidate to go away.
import "github.com/giantswarm/aws-operator/service/awsconfig/v2/cloudconfig"

// template related to this adapter: service/templates/cloudformation/guest/outputs.yaml

type outputsAdapter struct {
	MasterCloudConfigVersion string
	WorkerCloudConfigVersion string
}

func (o *outputsAdapter) getOutputs(cfg Config) error {
	o.MasterCloudConfigVersion = cloudconfig.MasterCloudConfigVersion
	o.WorkerCloudConfigVersion = cloudconfig.WorkerCloudConfigVersion

	return nil
}
