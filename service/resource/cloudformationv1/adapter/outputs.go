package adapter

import (
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
)

// template related to this adapter: service/templates/cloudformation/outputs.yaml

type outputsAdapter struct {
	ClusterVersion string
}

func (o *outputsAdapter) getOutputs(customObject awstpr.CustomObject, clients Clients) error {
	o.ClusterVersion = key.VersionBundleVersion(customObject)

	return nil
}
