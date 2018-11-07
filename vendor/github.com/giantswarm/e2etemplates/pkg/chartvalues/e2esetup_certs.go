package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type E2ESetupCertsConfig struct {
	Cluster      E2ESetupCertsConfigCluster
	CommonDomain string
}

type E2ESetupCertsConfigCluster struct {
	ID string
}

// NewE2ESetupCerts renders values required by e2esetup-certs-chart.
func NewE2ESetupCerts(config E2ESetupCertsConfig) (string, error) {
	if config.Cluster.ID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Cluster.ID must not be empty", config)
	}
	if config.CommonDomain == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}

	values, err := render.Render(e2eSetupCertsTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
