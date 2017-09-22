package microstorage

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
)

var (
	// RootKey may be used to list all values in the Storage.
	RootKey = K{key: "/"}
)

// K is an immutable, valid key.
type K struct {
	key string
}

// NewK creates a new immutable key to query the Storage with.
//
// The key should take a format of path separated by slashes "/". E.g.
// "/a/b/c". The key is sanitized before querying the storage. I.e. leading
// slash can be added and trailing slash can be removed. E.g. "/a/b/c",
// "a/b/c/", "a/b/c", and "/a/b/c/" represent the same key.
//
// NewK may fail if the key is not valid. See SanitizeKey godoc to learn how
// valid key looks like.
func NewK(key string) (K, error) {
	key, err := SanitizeKey(key)
	if err != nil {
		return K{}, microerror.Mask(err)
	}

	k := K{
		key: key,
	}
	return k, nil
}

// MustK is a helper that wraps a call to a function returning (K, error) and
// panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var k = microstorage.MustK(microstorage.NewK("key"))
func MustK(k K, err error) K {
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
	return k
}

// Key returns the actual sanitized key value.
func (k K) Key() string {
	return k.key
}

// KeyNoLeadingSlash returns the actual sanitized key value with leading slash
// stripped.
func (k K) KeyNoLeadingSlash() string {
	return k.key[1:]
}

// KV is an immutable key-value pair with valid key.
type KV struct {
	key string
	val string
}

// NewKV creates a new immutable key-value pair to update Storage with.
//
// The key should take a format of path separated by slashes "/". E.g.
// "/a/b/c". The key is sanitized before inserting to the storage. I.e. leading
// slash can be added and trailing slash can be removed. E.g. "/a/b/c",
// "a/b/c/", "a/b/c", and "/a/b/c/" represent the same key.
//
// The val is an arbitrary value stored under the key.
//
// NewKV may fail if the key is not valid. See SanitizeKey godoc to learn how
// valid key looks like.
func NewKV(key, val string) (KV, error) {
	key, err := SanitizeKey(key)
	if err != nil {
		return KV{}, microerror.Mask(err)
	}

	kv := KV{
		key: key,
		val: val,
	}
	return kv, nil
}

// MustKV is a helper that wraps a call to a function returning (KV, error) and
// panics if the error is non-nil. It is intended for use Storage
// implementations, where the key is known to be valid because it is retrieved
// from the storage.
func MustKV(kv KV, err error) KV {
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
	return kv
}

// K returns K instance created from the key value associated with this
// key-value pair.
func (k KV) K() K {
	return MustK(NewK(k.key))
}

// Key returns the sanitized key associated with this key-value pair.
func (k KV) Key() string {
	return k.key
}

// KeyNoLeadingSlashreturns the sanitized key associated with this key-value
// pair with leading slash stripped.
func (k KV) KeyNoLeadingSlash() string {
	return k.key[1:]
}

// Val returns the value associated with this key-value pair.
func (k KV) Val() string {
	return k.val
}

// Storage represents the abstraction for underlying storage backends.
type Storage interface {
	// Put stores the given value under the given key. If the value
	// under the key already exists Put overrides it.
	Put(ctx context.Context, kv KV) error
	// Delete removes the value stored under the given key.
	Delete(ctx context.Context, key K) error
	// Exists checks if a value under the given key exists or not.
	Exists(ctx context.Context, key K) (bool, error)
	// List does a lookup for all keys stored under the key, and returns the
	// relative key path, if any.
	// E.g: listing /foo/, with the key /foo/bar, returns bar.
	List(ctx context.Context, key K) ([]KV, error)
	// Search does a lookup for the value stored under key and returns it, if any.
	Search(ctx context.Context, key K) (KV, error)
}
