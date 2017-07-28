package kubeadmtokentpr

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List represents a list of custom objects.
type List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*CustomObject `json:"items" yaml:"items"`
}
