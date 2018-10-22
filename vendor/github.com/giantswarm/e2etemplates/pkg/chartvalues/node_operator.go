package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type NodeOperatorConfig struct {
	Namespace          string
	RegistryPullSecret string
}

func NewNodeOperator(config NodeOperatorConfig) (string, error) {
	if config.Namespace == "" {
		config.Namespace = "giantswarm"
	}
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}

	values, err := render.Render(nodeOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
