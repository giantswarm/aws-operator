package awstpr

import (
	"encoding/json"

	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"
)

// CustomObject represents the AWS TPR's custom object. It holds the
// specifications of the resource the AWS operator is interested in.
type CustomObject struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             api.ObjectMeta `json:"metadata"`
	Spec                 Spec           `json:"spec" yaml:"spec"`
}

func (co *CustomObject) GetObjectMeta() meta.Object {
	return &co.Metadata
}

func (co *CustomObject) GetObjectKind() unversioned.ObjectKind {
	return &co.TypeMeta
}

type coCopy CustomObject

func (co *CustomObject) UnmarshalJSON(data []byte) error {
	tmp := coCopy{}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return microerror.MaskAny(err)
	}

	tmp2 := CustomObject(tmp)
	*co = tmp2

	return nil
}
