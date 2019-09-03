package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/templates/cloudconfig"
)

// NOTE that code below is deprecated and needs refactoring.

func CloudConfigSmallTemplates() []string {
	return []string{
		cloudconfig.Small,
	}
}

func StatusAWSConfigNetworkCIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Status.Cluster.Network.CIDR
}
