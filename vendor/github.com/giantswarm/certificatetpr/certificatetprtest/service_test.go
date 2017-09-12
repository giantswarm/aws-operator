package certificatetprtest

import (
	"testing"

	"github.com/giantswarm/certificatetpr"
)

func Test_Fake_Service_SearcherInterface(t *testing.T) {
	newService := NewService()
	_, ok := interface{}(newService).(certificatetpr.Searcher)
	if !ok {
		t.Fatal("fake service does not implement Searcher interface")
	}
}
