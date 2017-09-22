package microstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK_NewKInvalid(t *testing.T) {
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

func TestK_KeyNoLeadingSlash(t *testing.T) {
	keys := []string{
		"a/b/c",
		"/a/b/c",
		"a/b/c/",
		"/a/b/c/",
	}
	for _, key := range keys {
		k := MustK(NewK(key))
		assert.Equal(t, "/a/b/c", k.Key(), "key=%s", key)
		assert.Equal(t, "a/b/c", k.KeyNoLeadingSlash(), "key=%s", key)
	}
}

func TestKV_NewKVInvalid(t *testing.T) {
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

func TestKV_KeyNoLeadingSlash(t *testing.T) {
	keys := []string{
		"a/b/c",
		"/a/b/c",
		"a/b/c/",
		"/a/b/c/",
	}
	for _, key := range keys {
		kv := MustKV(NewKV(key, "any-test-value"))
		assert.Equal(t, "/a/b/c", kv.Key(), "key=%s", key)
		assert.Equal(t, "a/b/c", kv.KeyNoLeadingSlash(), "key=%s", key)
	}
}
