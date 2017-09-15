package microstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKInvalid(t *testing.T) {
	keys := []string{
		"",
		"/",
		"//",
		"///",
		"////",
		"key//",
		"//key",
		"//key/",
		"/key//",
		"/key//",
		"in//between",
		"in///////between",
	}

	for _, key := range keys {
		_, err := NewK(key)
		assert.NotNil(t, err, "key=%s", key)
		assert.True(t, IsInvalidKey(err), "expected InvalidKeyError for key=%s", key)
	}
}

func TestNewKVInvalid(t *testing.T) {
	keys := []string{
		"",
		"/",
		"//",
		"///",
		"////",
		"key//",
		"//key",
		"//key/",
		"/key//",
		"/key//",
		"in//between",
		"in///////between",
	}

	for _, key := range keys {
		_, err := NewKV(key, "test-value")
		assert.NotNil(t, err, "key=%s", key)
		assert.True(t, IsInvalidKey(err), "expected InvalidKeyError for key=%s", key)
	}
}
