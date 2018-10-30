package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type E2ESetupVaultConfig struct {
	Vault E2ESetupVaultConfigVault
}

type E2ESetupVaultConfigVault struct {
	Token string
}

// NewE2ESetupVault renders values required by e2esetup-vault-chart.
func NewE2ESetupVault(config E2ESetupVaultConfig) (string, error) {
	if config.Vault.Token == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Vault.Token must not be empty", config)
	}

	values, err := render.Render(e2eSetupVaultTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
