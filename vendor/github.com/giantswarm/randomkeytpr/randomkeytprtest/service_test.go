package randomkeytprtest

import (
	"testing"

	"github.com/giantswarm/randomkeytpr"
)

func Test_Fake_Service_SearcherInterface(t *testing.T) {
	newService := NewService()
	_, ok := interface{}(newService).(randomkeytpr.Searcher)
	if !ok {
		t.Fatal("fake service does not implement Searcher interface")
	}
}
