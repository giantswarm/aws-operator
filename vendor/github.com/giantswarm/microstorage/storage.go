package microstorage

import (
	"context"
)

// Storage represents the abstraction for underlying storage backends. A storage
// implementation does not care about specific data types. All the
// storage cares about are key-value pairs. Code making use of the storage
// have to take care about specific types they care about them self.
type Storage interface {
	// Create is deprecated in favour of Put. Its semantics is unspecified
	// when the value of the key does not exist.
	//
	// Create stores the given value under the given key. Keys and values
	// might have specific schemes depending on the specific storage
	// implementation.  E.g. an etcd storage implementation will allow keys
	// to be defined as paths: path/to/key. Values might be JSON strings in
	// case the storage implementation wants to store its data as JSON
	// strings.
	Create(ctx context.Context, key, value string) error
	// Put stores the given value under the given key. If the value
	// under the key already exists Put overrides it. Keys and values might
	// have specific schemes depending on the specific storage
	// implementation.  E.g. an etcd storage implementation will allow keys
	// to be defined as paths: path/to/key. Values might be JSON strings in
	// case the storage implementation wants to store its data as JSON
	// strings.
	Put(ctx context.Context, key, value string) error
	// Delete removes the value stored under the given key.
	Delete(ctx context.Context, key string) error
	// Exists checks if a value under the given key exists or not.
	Exists(ctx context.Context, key string) (bool, error)
	// List does a lookup for all keys stored under the key, and returns the
	// relative key path, if any.
	// E.g: listing /foo/, with the key /foo/bar, returns bar.
	List(ctx context.Context, key string) ([]string, error)
	// Search does a lookup for the value stored under key and returns it, if any.
	Search(ctx context.Context, key string) (string, error)
}
