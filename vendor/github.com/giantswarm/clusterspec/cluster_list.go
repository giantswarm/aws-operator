package clusterspec

import "k8s.io/client-go/pkg/api/unversioned"

// TODO comment
type ClusterList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*Cluster `json:"items"`
}
