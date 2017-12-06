package adapter

import (
	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/awstpr"
)

// template related to this adapter: service/templates/cloudformation/outputs.yaml

type outputsAdapter struct {
	ClusterVersion string
}

func (o *outputsAdapter) getOutputs(customObject awstpr.CustomObject, clients Clients) error {
	o.ClusterVersion = keyv1.ClusterVersion(customObject)

	return nil
}
