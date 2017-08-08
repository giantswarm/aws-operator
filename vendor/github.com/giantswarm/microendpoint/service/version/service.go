package version

import (
	"context"
	"runtime"

	"github.com/giantswarm/microerror"
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Settings.
	Description string
	GitCommit   string
	Name        string
	Source      string
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Settings.
		Description: "",
		GitCommit:   "",
		Name:        "",
		Source:      "",
	}
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Settings.
	if config.Description == "" {
		return nil, microerror.Maskf(invalidConfigError, "description commit must not be empty")
	}
	if config.GitCommit == "" {
		return nil, microerror.Maskf(invalidConfigError, "git commit must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "name must not be empty")
	}
	if config.Source == "" {
		return nil, microerror.Maskf(invalidConfigError, "source must not be empty")
	}

	newService := &Service{
		Config: config,
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	Config
}

// Get returns the version response.
func (s *Service) Get(ctx context.Context, request Request) (*Response, error) {
	response := DefaultResponse()

	response.Description = s.Description
	response.GitCommit = s.GitCommit
	response.GoVersion = runtime.Version()
	response.Name = s.Name
	response.OSArch = runtime.GOOS + "/" + runtime.GOARCH
	response.Source = s.Source

	return response, nil
}
