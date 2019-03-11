package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type APIExtensionsFlannelConfigE2EConfig struct {
	ClusterID string
	Network   string
	VNI       int
}

func NewAPIExtensionsFlannelConfigE2E(config APIExtensionsFlannelConfigE2EConfig) (string, error) {
	if config.ClusterID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.Network == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Network must not be empty", config)
	}
	if config.VNI < 0 {
		return "", microerror.Maskf(invalidConfigError, "%T.VNI must not be negative number", config)
	}

	values, err := render.Render(apiExtensionsFlannelConfigE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
