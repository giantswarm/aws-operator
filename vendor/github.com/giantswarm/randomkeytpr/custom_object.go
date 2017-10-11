package randomkeytpr

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomObject represents the Random Key TPR's custom object. It holds the
// specifications of the resource the operator is interested in.
type CustomObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" yaml:"metadata"`

	Spec Spec `json:"spec" yaml:"spec"`
}
