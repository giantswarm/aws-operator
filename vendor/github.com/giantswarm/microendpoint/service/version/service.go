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
	Version        string
	VersionBundles []versionbundle.Bundle
}

// Service implements the version service interface.
type Service struct {
	description    string
	gitCommit      string
	name           string
	source         string
	version        string
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
	if config.Version == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Version must not be empty")
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
		version:        config.Version,
		versionBundles: config.VersionBundles,
	}

	newService.updateBuildInfoMetric()

	return newService, nil
}

// Get returns the version response.
func (s *Service) Get(ctx context.Context, request Request) (*Response, error) {

	response := &Response{
		Description:    s.description,
		GitCommit:      s.gitCommit,
		GoVersion:      runtime.Version(),
		Name:           s.name,
		OSArch:         runtime.GOOS + "/" + runtime.GOARCH,
		Source:         s.source,
		Version:        s.version,
		VersionBundles: s.versionBundles,
	}

	return response, nil
}
