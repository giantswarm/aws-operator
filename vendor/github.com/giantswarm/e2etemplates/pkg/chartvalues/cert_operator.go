package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type CertOperatorConfig struct {
	ClusterRole        CertOperatorConfigClusterRole
	ClusterRolePSP     CertOperatorConfigClusterRole
	CommonDomain       string
	CRD                CertOperatorConfigCRD
	Namespace          string
	RegistryPullSecret string
	PSP                CertOperatorPSP
	Vault              CertOperatorVault
}

type CertOperatorConfigClusterRole struct {
	BindingName string
	Name        string
}

type CertOperatorConfigCRD struct {
	// LabelSelector configures the operator's list watcher label selector to
	// consider only specific CRs. This is done e.g. for the kvm-operator e2e
	// tests due to the lack of test env encapsulation. Note that this option
	// therefore is optional.
	LabelSelector string
}

type CertOperatorPSP struct {
	Name string
}

type CertOperatorVault struct {
	Token string
}

func NewCertOperator(config CertOperatorConfig) (string, error) {
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
