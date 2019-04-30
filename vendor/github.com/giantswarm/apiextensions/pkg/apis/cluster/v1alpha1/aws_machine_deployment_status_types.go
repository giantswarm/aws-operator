package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSMachineDeploymentStatus is the structure put into the provider status of
// the Cluster API's MachineDeployment type. There it is tracked as serialized
// raw extension.
//
//     kind: AWSMachineDeploymentStatus
//     apiVersion: cluster.giantswarm.io/v1alpha1
//     metadata:
//       name: al9qy
//     cluster:
//       id: 8y5kc
//     nodePool:
//       id: al9qy
//       scaling:
//         desiredCapacity: 3
//       versions:
//       - lastTransitionTime: "2019-03-25T17:10:09.995948706Z"
//         version: 4.9.0
//     provider:
//       autoScalingGroup:
//         name: cluster-8y5kc-guest-main-workerAutoScalingGroup-1G3V6VQHBPY4O
//       availabilityZones:
//       - name: eu-central-1a
//         subnet:
//           private:
//             cidr: 10.1.6.0/25
//           public:
//             cidr: 10.1.6.128/25
//
type AWSMachineDeploymentStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Cluster           AWSMachineDeploymentStatusCluster  `json:"cluster" yaml:"cluster"`
	NodePool          AWSMachineDeploymentStatusNodePool `json:"nodePool" yaml:"nodePool"`
	Provider          AWSMachineDeploymentStatusProvider `json:"provider" yaml:"provider"`
}

type AWSMachineDeploymentStatusCluster struct {
	ID string `json:"id" yaml:"id"`
}

type AWSMachineDeploymentStatusNodePool struct {
	ID       string                                      `json:"id" yaml:"id"`
	Scaling  AWSMachineDeploymentStatusNodePoolScaling   `json:"scaling" yaml:"scaling"`
	Versions []AWSMachineDeploymentStatusNodePoolVersion `json:"versions" yaml:"versions"`
}

type AWSMachineDeploymentStatusNodePoolScaling struct {
	DesiredCapacity int `json:"desiredCapacity" yaml:"desiredCapacity"`
}

type AWSMachineDeploymentStatusNodePoolVersion struct {
	LastTransitionTime DeepCopyTime `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Version            string       `json:"version" yaml:"version"`
}

type AWSMachineDeploymentStatusProvider struct {
	AutoScalingGroup  AWSMachineDeploymentStatusProviderAutoScalingGroup   `json:"autoScalingGroup" yaml:"autoScalingGroup"`
	AvailabilityZones []AWSMachineDeploymentStatusProviderAvailabilityZone `json:"availabilityZones" yaml:"availabilityZones"`
}

type AWSMachineDeploymentStatusProviderAutoScalingGroup struct {
	Name string `json:"name" yaml:"name"`
}

type AWSMachineDeploymentStatusProviderAvailabilityZone struct {
	Name   string                                                   `json:"name" yaml:"name"`
	Subnet AWSMachineDeploymentStatusProviderAvailabilityZoneSubnet `json:"subnet" yaml:"subnet"`
}

type AWSMachineDeploymentStatusProviderAvailabilityZoneSubnet struct {
	Private AWSMachineDeploymentStatusProviderAvailabilityZoneSubnetPrivate `json:"private" yaml:"private"`
	Public  AWSMachineDeploymentStatusProviderAvailabilityZoneSubnetPublic  `json:"public" yaml:"public"`
}

type AWSMachineDeploymentStatusProviderAvailabilityZoneSubnetPrivate struct {
	CIDR string `json:"cidr" yaml:"cidr"`
}

type AWSMachineDeploymentStatusProviderAvailabilityZoneSubnetPublic struct {
	CIDR string `json:"cidr" yaml:"cidr"`
}
