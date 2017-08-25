package storagetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidKeyVariations(t *testing.T) {
	{
		old := validKeyVariationsIDGen
		defer func() {
			validKeyVariationsIDGen = old
		}()
	}

	keys := []string{
		"testkey",
		"/testkey",
		"testkey/",
		"/testkey/",
	}

	for _, key := range keys {
		validKeyVariationsIDGen = 8

		got := validKeyVariations(key)
		want := []string{
			"testkey-0009",
			"/testkey-0010",
			"testkey-0011/",
			"/testkey-0012/",
		}
		assert.Equal(t, want, got, "key=%s", key)
	}

}
