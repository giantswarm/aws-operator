package k8s

import (
	"encoding/json"

	"github.com/giantswarm/awstpr"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
)

type ClusterDecoder struct {
	decoder *json.Decoder
	close   func() error
}

func NewClusterDecoder(decoder *json.Decoder, closeFunc func() error) *ClusterDecoder {
	return &ClusterDecoder{
		decoder: decoder,
		close:   closeFunc,
	}
}

func (cd *ClusterDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object awstpr.CustomObject
	}
	if err := cd.decoder.Decode(&e); err != nil {
		return watch.Error, nil, err
	}

	return e.Type, &e.Object, nil
}

func (cd *ClusterDecoder) Close() {
	cd.close()
}
