package awstpr

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomObject represents the AWS TPR's custom object. It holds the
// specifications of the resource the AWS operator is interested in.
type CustomObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              Spec `json:"spec" yaml:"spec"`
}
