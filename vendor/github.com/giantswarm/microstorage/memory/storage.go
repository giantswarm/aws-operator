// Package memory provides a memory storage implementation.
package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microstorage"
)

// Config represents the configuration used to create a memory backed storage.
type Config struct {
}

// DefaultConfig provides a default configuration to create a new memory backed
// storage by best effort.
func DefaultConfig() Config {
	return Config{}
}

// New creates a new configured memory storage.
func New(config Config) (*Storage, error) {
	storage := &Storage{
		data:  map[string]string{},
		mutex: sync.Mutex{},
	}

	return storage, nil
}

// Storage is the memory backed storage.
type Storage struct {
	// Internals.

	data  map[string]string
	mutex sync.Mutex
}

func (s *Storage) Put(ctx context.Context, kv microstorage.KV) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[kv.Key()] = kv.Val()

	return nil
}

func (s *Storage) Delete(ctx context.Context, k microstorage.K) error {
	key := k.Key()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)

	return nil
}

func (s *Storage) Exists(ctx context.Context, k microstorage.K) (bool, error) {
	key := k.Key()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.data[key]

	return ok, nil
}

func (s *Storage) List(ctx context.Context, k microstorage.K) ([]microstorage.KV, error) {
	key := k.Key()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Special case.
	if key == "/" {
		var list []microstorage.KV
		for k, v := range s.data {
			k = k[1:] // append a key without leading '/'.
			list = append(list, microstorage.MustKV(microstorage.NewKV(k, v)))
		}
		return list, nil
	}

	var list []microstorage.KV

	i := len(key)
	for k, v := range s.data {
		if len(k) <= i+1 {
			continue
		}
		if !strings.HasPrefix(k, key) {
			continue
		}

		if k[i] != '/' {
			// We want to ignore all keys that are not separated by slash. When there
			// is a key stored like "foo/bar/baz", listing keys using "foo/ba" should
			// not succeed.
			continue
		}

		k = k[i+1:]
		list = append(list, microstorage.MustKV(microstorage.NewKV(k, v)))
	}

	return list, nil
}

func (s *Storage) Search(ctx context.Context, k microstorage.K) (microstorage.KV, error) {
	key := k.Key()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, ok := s.data[key]
	if ok {
		return microstorage.MustKV(microstorage.NewKV(key, value)), nil
	}

	return microstorage.KV{}, microerror.Maskf(microstorage.NotFoundError, "key=%s", key)
}
