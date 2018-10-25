package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type CertOperatorConfig struct {
	ClusterName        string
	ClusterRole        CertOperatorClusterRole
	ClusterRolePSP     CertOperatorClusterRole
	CommonDomain       string
	Namespace          string
	RegistryPullSecret string
	PSP                CertOperatorPSP
	Vault              CertOperatorVault
}

type CertOperatorClusterRole struct {
	BindingName string
	Name        string
}

type CertOperatorPSP struct {
	Name string
}

type CertOperatorVault struct {
	Token string
}

func NewCertOperator(config CertOperatorConfig) (string, error) {
	if config.ClusterName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterName must not be empty", config)
	}
	if config.ClusterRole.BindingName == "" {
		config.ClusterRole.BindingName = "cert-operator"
	}
	if config.ClusterRole.Name == "" {
		config.ClusterRole.Name = "cert-operator"
	}
	if config.ClusterRolePSP.BindingName == "" {
		config.ClusterRolePSP.BindingName = "cert-operator-psp"
	}
	if config.ClusterRolePSP.Name == "" {
		config.ClusterRolePSP.Name = "cert-operator-psp"
	}
	if config.CommonDomain == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}
	if config.Namespace == "" {
		config.Namespace = "giantswarm"
	}
	if config.PSP.Name == "" {
		config.PSP.Name = "cert-operator-psp"
	}
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}
	if config.Vault.Token == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Vault.Token must not be empty", config)
	}

	values, err := render.Render(certOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
