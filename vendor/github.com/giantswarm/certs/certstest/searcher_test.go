package certstest

import (
	"testing"

	"github.com/giantswarm/certs"
)

func Test_CertsTest_NewSearcher(t *testing.T) {
	s := NewSearcher()
	_, ok := interface{}(s).(certs.Interface)
	if !ok {
		t.Fatal("searcher does not implement correct interface")
	}
}
