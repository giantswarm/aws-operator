package k8s

import (
	"encoding/json"
	"io"

	"github.com/giantswarm/clusterspec"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
)

type ClusterDecoder struct {
	Stream io.ReadCloser
}

func (d *ClusterDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	decoder := json.NewDecoder(d.Stream)

	var e struct {
		Type   watch.EventType
		Object clusterspec.Cluster
	}
	if err := decoder.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}

func (d *ClusterDecoder) Close() {
	d.Stream.Close()
}
