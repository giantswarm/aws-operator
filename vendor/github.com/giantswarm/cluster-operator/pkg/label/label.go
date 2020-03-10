// Package label contains common Kubernetes metadata. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// Certificate label identifies certificates in CertConfig CRs.
	Certificate = "giantswarm.io/certificate"
	// Cluster label is a new style label for ClusterID
	Cluster = "giantswarm.io/cluster"
	// MachineDeployment label denotes which node pool corresponding resources
	// belongs.
	MachineDeployment = "giantswarm.io/machine-deployment"
	// ManagedBy label denotes which operator manages corresponding resource.
	ManagedBy = "giantswarm.io/managed-by"
	// Organization label denotes guest cluster's organization ID as displayed
	// in the front-end.
	Organization = "giantswarm.io/organization"
)
