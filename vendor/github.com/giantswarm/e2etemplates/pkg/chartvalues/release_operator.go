package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type ReleaseOperatorConfig struct {
	RegistryPullSecret string
}

// NewAWSOperator renders values required by aws-operator-chart.
func NewReleaseOperator(config ReleaseOperatorConfig) (string, error) {
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}

	values, err := render.Render(releaseOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
