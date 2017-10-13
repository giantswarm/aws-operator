package metricsstorage

import (
	"testing"

	"github.com/giantswarm/microstorage/memory"
	"github.com/giantswarm/microstorage/storagetest"
)

func TestMetricsStorage(t *testing.T) {
	underlying, err := memory.New(memory.DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
	}

	config := DefaultConfig()
	config.Underlying = underlying

	storage, err := New(config)
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
	}

	storagetest.Test(t, storage)
}
