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

func (s *Storage) Create(ctx context.Context, key, value string) error {
	var err error

	key, err = microstorage.SanitizeKey(key)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.Put(ctx, key, value)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (s *Storage) Put(ctx context.Context, key, value string) error {
	var err error

	key, err = microstorage.SanitizeKey(key)
	if err != nil {
		return microerror.Mask(err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value

	return nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	var err error

	key, err = microstorage.SanitizeKey(key)
	if err != nil {
		return microerror.Mask(err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)

	return nil
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	var err error

	key, err = microstorage.SanitizeKey(key)
	if err != nil {
		return false, microerror.Mask(err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.data[key]

	return ok, nil
}

func (s *Storage) List(ctx context.Context, key string) ([]string, error) {
	var err error

	key, err = microstorage.SanitizeListKey(key)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Special case.
	if key == "/" {
		var list []string
		for k, _ := range s.data {
			// append a key without leading '/'.
			list = append(list, k[1:])
		}
		return list, nil
	}

	var list []string

	i := len(key)
	for k, _ := range s.data {
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

		list = append(list, k[i+1:])
	}

	return list, nil
}

func (s *Storage) Search(ctx context.Context, key string) (string, error) {
	var err error

	key, err = microstorage.SanitizeKey(key)
	if err != nil {
		return "", microerror.Mask(err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, ok := s.data[key]
	if ok {
		return value, nil
	}

	return "", microerror.Maskf(microstorage.NotFoundError, key)
}
