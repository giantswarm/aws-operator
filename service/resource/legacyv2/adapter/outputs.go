package adapter

import (
	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/outputs.yaml

type outputsAdapter struct {
	ClusterVersion string
}

func (o *outputsAdapter) getOutputs(cfg Config) error {
	o.ClusterVersion = keyv2.ClusterVersion(cfg.CustomObject)

	return nil
}
