package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type APIExtensionsReleaseE2EConfig struct {
	Active        bool
	Authorities   []APIExtensionsReleaseE2EConfigAuthority
	Date          string
	Name          string
	Namespace     string
	Provider      string
	Version       string
	VersionBundle APIExtensionsReleaseE2EConfigVersionBundle
}

type APIExtensionsReleaseE2EConfigAuthority struct {
	Name    string
	Version string
}

type APIExtensionsReleaseE2EConfigVersionBundle struct {
	Version string
}

// NewAPIExtensionsAWSConfigE2E renders values required by apiextensions-aws-config-e2e-chart.
func NewAPIExtensionsReleaseE2E(config APIExtensionsReleaseE2EConfig) (string, error) {
	if config.Date == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Date must not be empty", config)
	}
	if config.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Name must not be empty", config)
	}
	if config.Namespace == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}
	if config.Provider == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Version must not be empty", config)
	}

	if len(config.Authorities) == 0 {
		return "", microerror.Maskf(invalidConfigError, "%T.Authorities must not be empty", config)
	}
	if config.VersionBundle.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.VersionBundle.Version must not be empty", config)
	}

	values, err := render.Render(apiExtensionsReleaseE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
