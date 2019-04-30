package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func ClusterAPIEndpoint(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Kubernetes.API.Domain
}

func ClusterID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.ID
}

func ClusterNamespace(customObject v1alpha1.AWSConfig) string {
	return ClusterID(customObject)
}

func BaseDomain(customObject v1alpha1.AWSConfig) string {
	// TODO remove other zones and make it a BaseDomain in the CR.
	// CloudFormation creates a separate HostedZone with the same name.
	// Probably the easiest way for now is to just allow single domain for
	// everything which we do now.
	return customObject.Spec.AWS.HostedZones.API.Name
}
