package retrystorage

import (
	"testing"

	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage/memory"
	"github.com/giantswarm/microstorage/storagetest"
)

func TestRetryStorage(t *testing.T) {
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
	}

	underlying, err := memory.New(memory.DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
	}

	config := DefaultConfig()
	config.Logger = logger
	config.Underlying = underlying

	storage, err := New(config)
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
	}

	storagetest.Test(t, storage)
}
