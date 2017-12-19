package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/outputs.yaml

type outputsAdapter struct {
	ClusterVersion string
}

func (o *outputsAdapter) getOutputs(customObject v1alpha1.AWSConfig, clients Clients) error {
	o.ClusterVersion = keyv2.ClusterVersion(customObject)

	return nil
}
