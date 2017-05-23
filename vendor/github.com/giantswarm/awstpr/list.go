package awstpr

import (
	"encoding/json"

	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api/unversioned"
)

type List struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*CustomObject `json:"items" yaml:"items"`
}

func (l *List) GetListMeta() unversioned.List {
	return &l.Metadata
}

func (l *List) GetObjectKind() unversioned.ObjectKind {
	return &l.TypeMeta
}

type lCopy List

func (l *List) UnmarshalJSON(data []byte) error {
	tmp := lCopy{}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return microerror.MaskAny(err)
	}

	tmp2 := List(tmp)
	*l = tmp2

	return nil
}
