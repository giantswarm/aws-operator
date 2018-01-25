package crdstorage

import (
	"testing"

	"github.com/giantswarm/microstorage"
)

func TestInterface(t *testing.T) {
	var _ microstorage.Storage = &Storage{}
}
