package microstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"

	"github.com/giantswarm/microendpoint/service/healthz"
)

const (
	// Description describes which functionality this health check implements.
	Description = "Ensure microstorage availability."
	// Name is the identifier of the health check. This can be used for emitting
	// metrics.
	Name = "microstorage"
	// SuccessMessage is the message returned in case the health check did not
	// fail.
	SuccessMessage = "all good"
	// Timeout is the time being waited until timing out health check, which
	// renders its result unsuccessful.
	Timeout = 10 * time.Second
)

const (
	HealthCheckKey   = "microstorage-health-check-key"
	HealthCheckValue = "microstorage-health-check-value"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Storage

	// Settings.
	Timeout time.Duration
}

// DefaultConfig provides a default configuration to create a new healthz service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:  nil,
		Storage: nil,

		// Settings.
		Timeout: Timeout,
	}
}

// Service implements the healthz service interface.
type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Storage

	// Settings.
	timeout time.Duration
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Storage must not be empty")
	}

	// Settings.
	if config.Timeout.Seconds() == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.Timout must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,

		// Settings.
		timeout: config.Timeout,
	}

	return newService, nil
}

// GetHealthz implements the health check for Kubernetes.
func (s *Service) GetHealthz(ctx context.Context) (healthz.Response, error) {
	failed := false
	message := SuccessMessage
	{
		ch := make(chan string, 1)

		go func() {
			err := s.getHealthzWithError(ctx)
			if err != nil {
				ch <- err.Error()
				return
			}
			ch <- ""
		}()

		select {
		case m := <-ch:
			if m != "" {
				failed = true
				message = m
			}
		case <-time.After(s.timeout):
			failed = true
			message = fmt.Sprintf("timed out after %s", s.timeout)
		}
	}

	response := healthz.Response{
		Description: Description,
		Failed:      failed,
		Message:     message,
		Name:        Name,
	}

	return response, nil
}

func (s *Service) getHealthzWithError(ctx context.Context) error {
	err := s.storage.Put(ctx, HealthCheckKey, HealthCheckValue)
	if err != nil {
		return microerror.Mask(err)
	}

	v, err := s.storage.Search(ctx, HealthCheckKey)
	if err != nil {
		return microerror.Mask(err)
	}
	if v != HealthCheckValue {
		return microerror.Maskf(executionFailedError, "expected health check value '%s' got '%s'", HealthCheckValue, v)
	}

	return nil
}
