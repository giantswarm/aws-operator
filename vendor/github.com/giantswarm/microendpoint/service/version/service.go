package version

import (
	"context"
	"runtime"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Settings.
	Description    string
	GitCommit      string
	Name           string
	Source         string
	VersionBundles []versionbundle.Bundle
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Settings.
		Description:    "",
		GitCommit:      "",
		Name:           "",
		Source:         "",
		VersionBundles: nil,
	}
}

// Service implements the version service interface.
type Service struct {
	description    string
	gitCommit      string
	name           string
	source         string
	versionBundles []versionbundle.Bundle
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Settings.
	if config.Description == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Description must not be empty")
	}
	if config.GitCommit == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GitCommit must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}
	if config.Source == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Source must not be empty")
	}

	if len(config.VersionBundles) != 0 {
		err := versionbundle.Bundles(config.VersionBundles).Validate()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		description:    config.Description,
		gitCommit:      config.GitCommit,
		name:           config.Name,
		source:         config.Source,
		versionBundles: config.VersionBundles,
	}

	return newService, nil
}

// Get returns the version response.
func (s *Service) Get(ctx context.Context, request Request) (*Response, error) {
	response := DefaultResponse()

	response.Description = s.description
	response.GitCommit = s.gitCommit
	response.GoVersion = runtime.Version()
	response.Name = s.name
	response.OSArch = runtime.GOOS + "/" + runtime.GOARCH
	response.Source = s.source
	response.VersionBundles = s.versionBundles

	return response, nil
}
