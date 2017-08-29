package retrystorage

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
)

const (
	// defaultMaxAttempts is the default number of attempts of each storate
	// operations before it finally fails.
	defaultMaxAttempts = 3
)

type stoppingBackOff struct {
	attempts    int
	MaxAttempts int
	Underlying  backoff.BackOff
}

func (s *stoppingBackOff) NextBackOff() time.Duration {
	if s.attempts >= s.MaxAttempts {
		return backoff.Stop
	}
	s.attempts++
	return s.Underlying.NextBackOff()
}

func (s *stoppingBackOff) Reset() {
	s.attempts = 0
	s.Underlying.Reset()
}

type Config struct {
	// Dependencies.

	Logger     micrologger.Logger
	Underlying microstorage.Storage

	// Settings.

	NewBackOffFunc func() backoff.BackOff
}

func DefaultConfig() Config {
	return Config{
		// Dependencies.

		Logger:     nil, // Required.
		Underlying: nil, // Required.

		// Settings.

		NewBackOffFunc: func() backoff.BackOff {
			return &stoppingBackOff{
				MaxAttempts: defaultMaxAttempts,
				Underlying:  backoff.NewExponentialBackOff(),
			}
		},
	}
}

type Storage struct {
	logger         micrologger.Logger
	underlying     microstorage.Storage
	newBackOffFunc func() backoff.BackOff
}

func New(config Config) (*Storage, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger is empty")
	}
	if config.Underlying == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Underlying is empty")
	}
	if config.NewBackOffFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.NewBackOffFunc is empty")
	}

	s := &Storage{
		logger:         config.Logger,
		underlying:     config.Underlying,
		newBackOffFunc: config.NewBackOffFunc,
	}

	return s, nil
}

func (s *Storage) Create(ctx context.Context, key, value string) error {
	b := s.newBackOffFunc()
	op := func() error {
		err := s.underlying.Create(ctx, key, value)
		if microstorage.IsInvalidKey(err) || microstorage.IsNotFound(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	notify := func(err error, delay time.Duration) {
		s.logger.Log("warning", "retrying", "op", "create", "key", key, "delay", delay, "err", fmt.Sprintf("%#v", err))
	}
	err := backoff.RetryNotify(op, b, notify)
	return microerror.Mask(err)
}

func (s *Storage) Put(ctx context.Context, key, value string) error {
	b := s.newBackOffFunc()
	op := func() error {
		err := s.underlying.Put(ctx, key, value)
		if microstorage.IsInvalidKey(err) || microstorage.IsNotFound(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	notify := func(err error, delay time.Duration) {
		s.logger.Log("warning", "retrying", "op", "put", "key", key, "delay", delay, "err", fmt.Sprintf("%#v", err))
	}
	err := backoff.RetryNotify(op, b, notify)
	return microerror.Mask(err)
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	b := s.newBackOffFunc()
	op := func() error {
		err := s.underlying.Delete(ctx, key)
		if microstorage.IsInvalidKey(err) || microstorage.IsNotFound(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	notify := func(err error, delay time.Duration) {
		s.logger.Log("warning", "retrying", "op", "delete", "key", key, "delay", delay, "err", fmt.Sprintf("%#v", err))
	}
	err := backoff.RetryNotify(op, b, notify)
	return microerror.Mask(err)
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	b := s.newBackOffFunc()
	var exists bool
	op := func() error {
		var err error
		exists, err = s.underlying.Exists(ctx, key)
		if microstorage.IsInvalidKey(err) || microstorage.IsNotFound(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	notify := func(err error, delay time.Duration) {
		s.logger.Log("warning", "retrying", "op", "exists", "key", key, "delay", delay, "err", fmt.Sprintf("%#v", err))
	}
	err := backoff.RetryNotify(op, b, notify)
	return exists, microerror.Mask(err)
}

func (s *Storage) List(ctx context.Context, key string) ([]string, error) {
	b := s.newBackOffFunc()
	var list []string
	op := func() error {
		var err error
		list, err = s.underlying.List(ctx, key)
		if microstorage.IsInvalidKey(err) || microstorage.IsNotFound(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	notify := func(err error, delay time.Duration) {
		s.logger.Log("warning", "retrying", "op", "list", "key", key, "delay", delay, "err", fmt.Sprintf("%#v", err))
	}
	err := backoff.RetryNotify(op, b, notify)
	return list, microerror.Mask(err)
}

func (s *Storage) Search(ctx context.Context, key string) (string, error) {
	b := s.newBackOffFunc()
	var value string
	op := func() error {
		var err error
		value, err = s.underlying.Search(ctx, key)
		if microstorage.IsInvalidKey(err) || microstorage.IsNotFound(err) {
			return backoff.Permanent(err)
		}
		return err
	}
	notify := func(err error, delay time.Duration) {
		s.logger.Log("warning", "retrying", "op", "search", "key", key, "delay", delay, "err", fmt.Sprintf("%#v", err))
	}
	err := backoff.RetryNotify(op, b, notify)
	return value, microerror.Mask(err)
}
