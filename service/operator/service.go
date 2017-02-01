package operator

import (
	"sync"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Foo string
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Settings.
		Foo: "",
	}
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	// Settings.
	if config.Foo == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "foo must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger: config.Logger,

		// Internals
		bootOnce: sync.Once{},

		// Settings.
		foo: config.Foo,
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	logger micrologger.Logger

	// Internals.
	bootOnce sync.Once

	// Settings.
	foo string
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.logger.Log("debug", "fix me, I do useless shit", "foo", s.foo)
	})
}
