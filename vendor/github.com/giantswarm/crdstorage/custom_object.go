package crdstorage

import (
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// customObject represents the crdstorage CRD's custom object. It holds the
// storage data.
type customObject struct {
	apismetav1.TypeMeta   `json:",inline"`
	apismetav1.ObjectMeta `json:"metadata,omitempty"`

	Data map[string]string `json:"data"`
}
