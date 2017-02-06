package clusterspec

import (
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

// TODO comment
type Cluster struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 ClusterSpec `json:"spec"`
}

// TODO comment
func (c *Cluster) GetObjectKind() unversioned.ObjectKind {
	return nil
}
