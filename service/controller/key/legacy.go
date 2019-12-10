package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

// NOTE that code below is deprecated and needs refactoring.

func StatusAWSConfigNetworkCIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Status.Cluster.Network.CIDR
}
