package randomkeystest

import (
	"testing"

	"github.com/giantswarm/randomkeys"
)

func Test_RandomReysTest_NewSearcher(t *testing.T) {
	s := NewSearcher()
	_, ok := interface{}(s).(randomkeys.Interface)
	if !ok {
		t.Fatal("searcher does not implement correct interface")
	}
}
