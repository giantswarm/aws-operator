package certificatetpr

import (
	"k8s.io/client-go/pkg/api/unversioned"
)

type List struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*CustomObject `json:"items"`
}
