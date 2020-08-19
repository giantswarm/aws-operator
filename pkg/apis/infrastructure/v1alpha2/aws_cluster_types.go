package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kindAWSCluster              = "AWSCluster"
	awsClusterDocumentationLink = "https://docs.giantswarm.io/reference/cp-k8s-api/awsclusters.infrastructure.giantswarm.io/"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=aws;cluster-api;giantswarm
// +k8s:openapi-gen=true

// AWSCluster is the infrastructure provider referenced in upstream CAPI Cluster
// CRs.
type AWSCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AWSClusterSpec `json:"spec"`
	// +kubebuilder:validation:Optional
	// Spec part of the AWSCluster resource.
	Status AWSClusterStatus `json:"status,omitempty"`
}

// AWSClusterSpec is the spec part for the AWSCluster resource.
// +k8s:openapi-gen=true
type AWSClusterSpec struct {
	// Cluster specification details.
	Cluster AWSClusterSpecCluster `json:"cluster"`
	// Provider-specific configuration details.
	Provider AWSClusterSpecProvider `json:"provider"`
}

// AWSClusterSpecCluster provides cluster specification details.
// +k8s:openapi-gen=true
type AWSClusterSpecCluster struct {
	// User-friendly description that should explain the purpose of the
	// cluster to humans.
	Description string `json:"description"`
	// DNS configuration details.
	DNS AWSClusterSpecClusterDNS `json:"dns"`
	// +kubebuilder:validation:Optional
	// Flags passed to kube-proxy on each node.
	KubeProxy AWSClusterSpecClusterKubeProxy `json:"kubeProxy,omitempty"`
	// Configuration for OpenID Connect (OIDC) authentication.
	OIDC AWSClusterSpecClusterOIDC `json:"oidc,omitempty"`
}

// AWSClusterSpecClusterDNS holds DNS configuration details.
// +k8s:openapi-gen=true
type AWSClusterSpecClusterDNS struct {
	Domain string `json:"domain"`
}

// AWSClusterSpecClusterOIDC holds configuration for OpenID Connect (OIDC) authentication.
// +k8s:openapi-gen=true
type AWSClusterSpecClusterOIDC struct {
	Claims    AWSClusterSpecClusterOIDCClaims `json:"claims,omitempty"`
	ClientID  string                          `json:"clientID,omitempty"`
	IssuerURL string                          `json:"issuerURL,omitempty"`
}

// AWSClusterSpecClusterOIDCClaims defines OIDC claims.
// +k8s:openapi-gen=true
type AWSClusterSpecClusterOIDCClaims struct {
	Username string `json:"username,omitempty"`
	Groups   string `json:"groups,omitempty"`
}

// AWSClusterSpecClusterKubeProxy describes values passed to the kube-proxy running in a tenant cluster.
// +k8s:openapi-gen=true
type AWSClusterSpecClusterKubeProxy struct {
	// Maximum number of NAT connections to track per CPU core (0 for default).
	// Passed to kube-proxy as --conntrack-max-per-core.
	ConntrackMaxPerCore int `json:"conntrackMaxPerCore,omitempty"`
}

// AWSClusterSpecProvider holds some AWS details.
// +k8s:openapi-gen=true
type AWSClusterSpecProvider struct {
	// Location of a secret providing the ARN of AWS IAM identity
	// to use with this cluster.
	CredentialSecret AWSClusterSpecProviderCredentialSecret `json:"credentialSecret"`
	// +kubebuilder:validation:Optional
	// Master holds master node configuration details.
	// Note that this attribute is being deprecated. The master node specification can now be found in the AWSControlPlane resource.
	Master AWSClusterSpecProviderMaster `json:"master,omitempty"`
	// +kubebuilder:validation:Optional
	// Pod network configuration.
	Pods AWSClusterSpecProviderPods `json:"pods,omitempty"`
	// AWS region the cluster is to be running in.
	Region string `json:"region"`
}

// AWSClusterSpecProviderCredentialSecret details how to chose the AWS IAM identity ARN
// to use with this cluster.
// +k8s:openapi-gen=true
type AWSClusterSpecProviderCredentialSecret struct {
	// Name of the provider credential resoure.
	Name string `json:"name"`
	// Kubernetes namespace holding the provider credential.
	Namespace string `json:"namespace"`
}

// AWSClusterSpecProviderMaster holds master node configuration details.
// +k8s:openapi-gen=true
type AWSClusterSpecProviderMaster struct {
	// +kubebuilder:validation:Optional
	// AWS availability zone to place the master node in.
	AvailabilityZone string `json:"availabilityZone"`
	// +kubebuilder:validation:Optional
	// AWS EC2 instance type to use for the master node.
	InstanceType string `json:"instanceType"`
}

// AWSClusterSpecProviderPods Pod network configuration.
// +k8s:openapi-gen=true
type AWSClusterSpecProviderPods struct {
	// +kubebuilder:validation:Optional
	// IPv4 address block used for pods, in CIDR notation.
	CIDRBlock string `json:"cidrBlock,omitempty"`
	// +kubebuilder:validation:Optional
	// When set to false, pod connections outside the VPC where the pod is located will be NATed through the node primary IP. When set to true, all connections will use the pod IP.
	ExternalSNAT *bool `json:"externalSNAT,omitempty"`
}

// AWSClusterStatus holds status information about the cluster, populated once the
// cluster is in creation or created.
// +k8s:openapi-gen=true
type AWSClusterStatus struct {
	// +kubebuilder:validation:Optional
	// Cluster-specific status details, including conditions and versions.
	Cluster CommonClusterStatus `json:"cluster,omitempty"`
	// +kubebuilder:validation:Optional
	// Provider-specific status details.
	Provider AWSClusterStatusProvider `json:"provider,omitempty"`
}

// AWSClusterStatusProvider holds provider-specific status details.
// +k8s:openapi-gen=true
type AWSClusterStatusProvider struct {
	// +kubebuilder:validation:Optional
	// Network-specific configuration details
	Network AWSClusterStatusProviderNetwork `json:"network,omitempty"`
}

// AWSClusterStatusProviderNetwork holds network details.
// +k8s:openapi-gen=true
type AWSClusterStatusProviderNetwork struct {
	// +kubebuilder:validation:Optional
	// IPv4 address block used by the tenant cluster nodes, in CIDR notation.
	CIDR string `json:"cidr,omitempty"`
	// +kubebuilder:validation:Optional
	// Identifier of the AWS Virtual Private Cloud (VPC) of the tenant cluster, e.g. `vpc-1234567890abcdef0`.
	VPCID string `json:"vpcID,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSClusterList is the type returned when listing AWSCLuster resources.
type AWSClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AWSCluster `json:"items"`
}
