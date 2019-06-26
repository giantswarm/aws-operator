package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSClusterStatus is the structure put into the provider status of the
// Cluster API's Cluster type. There it is tracked as serialized raw extension.
//
//     kind: AWSClusterStatus
//     apiVersion: cluster.giantswarm.io/v1alpha1
//     metadata:
//       name: 8y5kc
//     cluster:
//       conditions:
//       - lastTransitionTime: "2019-03-25T17:10:09.333633991Z"
//         type: Created
//       id: 8y5kc
//       versions:
//       - lastTransitionTime: "2019-03-25T17:10:09.995948706Z"
//         version: 4.9.0
//     provider:
//       network:
//         cidr: 10.1.6.0/24
//
type AWSClusterStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Cluster           CommonClusterStatus      `json:"cluster" yaml:"cluster"`
	Provider          AWSClusterStatusProvider `json:"provider" yaml:"provider"`
}

type AWSClusterStatusProvider struct {
	Network AWSClusterStatusProviderNetwork `json:"network" yaml:"network"`
}

type AWSClusterStatusProviderNetwork struct {
	CIDR string `json:"cidr" yaml:"cidr"`
}
