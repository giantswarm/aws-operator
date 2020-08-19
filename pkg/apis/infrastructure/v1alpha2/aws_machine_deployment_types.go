package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=aws;giantswarm;cluster-api
// +k8s:openapi-gen=true

// AWSMachineDeployment is the infrastructure provider referenced in Kubernetes Cluster API MachineDeployment resources.
// It contains provider-specific specification and status for a node pool.
// In use on AWS since Giant Swarm release v10.x.x and reconciled by aws-operator.
type AWSMachineDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Contains the specification.
	Spec AWSMachineDeploymentSpec `json:"spec"`
	// +kubebuilder:validation:Optional
	// Holds status information.
	Status AWSMachineDeploymentStatus `json:"status,omitempty"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpec struct {
	// Specifies details of node pool and the worker nodes it should contain.
	NodePool AWSMachineDeploymentSpecNodePool `json:"nodePool"`
	// Contains AWS specific details.
	Provider AWSMachineDeploymentSpecProvider `json:"provider"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpecNodePool struct {
	// User-friendly name or description of the purpose of the node pool.
	Description string `json:"description"`
	// Specification of the worker node machine.
	Machine AWSMachineDeploymentSpecNodePoolMachine `json:"machine"`
	// Scaling settings for the node pool, configuring the cluster-autoscaler
	// determining the number of nodes to have in this node pool.
	Scaling AWSMachineDeploymentSpecNodePoolScaling `json:"scaling"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpecNodePoolMachine struct {
	// Size of the volume reserved for Docker images and overlay file systems of
	// Docker containers. Unit: 1 GB = 1,000,000,000 Bytes.
	DockerVolumeSizeGB int `json:"dockerVolumeSizeGB"`
	// Size of the volume reserved for the kubelet, which can be used by Pods via
	// volumes of type EmptyDir. Unit: 1 GB = 1,000,000,000 Bytes.
	KubeletVolumeSizeGB int `json:"kubeletVolumeSizeGB"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpecNodePoolScaling struct {
	// Maximum number of worker nodes in this node pool.
	Max int `json:"max"`
	// Minimum number of worker nodes in this node pool.
	Min int `json:"min"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpecProvider struct {
	// Name(s) of the availability zone(s) to use for worker nodes. Using multiple
	// availability zones results in higher resilience but can also result in higher
	// cost due to network traffic between availability zones.
	AvailabilityZones []string `json:"availabilityZones"`
	// +kubebuilder:validation:Optional
	// Settings defining the distribution of on-demand and spot instances in the node pool.
	InstanceDistribution AWSMachineDeploymentSpecInstanceDistribution `json:"instanceDistribution,omitempty"`
	// Specification of worker nodes.
	Worker AWSMachineDeploymentSpecProviderWorker `json:"worker"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpecInstanceDistribution struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// Base capacity of on-demand instances to use for worker nodes in this pool. When this larger
	// than 0, this value defines a number of worker nodes that will be created using on-demand
	// EC2 instances, regardless of the value configured as `onDemandPercentageAboveBaseCapacity`.
	OnDemandBaseCapacity int `json:"onDemandBaseCapacity"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// Percentage of on-demand EC2 instances to use for worker nodes, instead of spot instances,
	// for instances exceeding `onDemandBaseCapacity`. For example, to have half of the worker nodes
	// use spot instances and half use on-demand, set this value to 50.
	OnDemandPercentageAboveBaseCapacity *int `json:"onDemandPercentageAboveBaseCapacity"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentSpecProviderWorker struct {
	// AWS EC2 instance type name to use for the worker nodes in this node pool.
	InstanceType string `json:"instanceType"`
	// +kubebuilder:default=false
	// If true, certain instance types with specs similar to instanceType will be used.
	UseAlikeInstanceTypes bool `json:"useAlikeInstanceTypes"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentStatus struct {
	// +kubebuilder:validation:Optional
	// Status specific to AWS.
	Provider AWSMachineDeploymentStatusProvider `json:"provider,omitempty"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentStatusProvider struct {
	// +kubebuilder:validation:Optional
	// Status of worker nodes.
	Worker AWSMachineDeploymentStatusProviderWorker `json:"worker,omitempty"`
}

// +k8s:openapi-gen=true
type AWSMachineDeploymentStatusProviderWorker struct {
	// +kubebuilder:validation:Optional
	// AWS EC2 instance types used for the worker nodes in this node pool.
	InstanceTypes []string `json:"instanceTypes,omitempty"`
	// +kubebuilder:validation:Optional
	// Number of EC2 spot instances used in this node pool.
	SpotInstances int `json:"spotInstances,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AWSMachineDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AWSMachineDeployment `json:"items"`
}
