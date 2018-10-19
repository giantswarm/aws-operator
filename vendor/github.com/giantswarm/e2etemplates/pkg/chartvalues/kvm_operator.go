package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type KVMOperatorConfig struct {
	ClusterName        string
	ClusterRole        KVMOperatorClusterRole
	ClusterRolePSP     KVMOperatorClusterRole
	Namespace          string
	PSP                KVMOperatorPSP
	RegistryPullSecret string
}

type KVMOperatorClusterRole struct {
	BindingName string
	Name        string
}

type KVMOperatorPSP struct {
	Name string
}

func NewKVMOperator(config KVMOperatorConfig) (string, error) {
	if config.ClusterName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterName must not be empty", config)
	}
	if config.ClusterRole.BindingName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRole.BindingName must not be empty", config)
	}
	if config.ClusterRole.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRole.Name must not be empty", config)
	}
	if config.ClusterRolePSP.BindingName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRolePSP.BindingName must not be empty", config)
	}
	if config.ClusterRolePSP.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRolePSP.Name must not be empty", config)
	}
	if config.Namespace == "" {
		config.Namespace = "giantswarm"
	}
	if config.PSP.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.PSP.Name must not be empty", config)
	}
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}

	values, err := render.Render(kvmOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
