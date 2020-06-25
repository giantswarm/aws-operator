package key

import (
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/internal/templates/cloudconfig"
)

// NOTE that code below is deprecated and needs refactoring.

func CloudConfigSmallTemplates() []string {
	return []string{
		cloudconfig.Small,
	}
}

func StatusAWSConfigNetworkCIDR(customObject g8sv1alpha1.AWSConfig) string {
	return customObject.Status.Cluster.Network.CIDR
}
