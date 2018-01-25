package legacytest

import (
	"testing"

	"github.com/giantswarm/certs/legacy"
)

func Test_Fake_Service_SearcherInterface(t *testing.T) {
	newService := NewService()
	_, ok := interface{}(newService).(legacy.Searcher)
	if !ok {
		t.Fatal("fake service does not implement Searcher interface")
	}
}
