package awstpr

import (
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

// CustomObject represents the AWS TPR's custom object. It holds the
// specifications of the resource the AWS operator is interested in.
type CustomObject struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             v1.ObjectMeta `json:"metadata,omitempty"`
	Spec                 Spec          `json:"spec" yaml:"spec"`
}

func (co *CustomObject) GetObjectMeta() meta.Object {
	return &co.Metadata
}

func (co *CustomObject) GetObjectKind() unversioned.ObjectKind {
	return &co.TypeMeta
}
