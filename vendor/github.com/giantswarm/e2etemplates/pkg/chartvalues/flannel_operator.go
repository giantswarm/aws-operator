package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type FlannelOperatorConfig struct {
	ClusterName        string
	ClusterRole        FlannelOperatorClusterRole
	ClusterRolePSP     FlannelOperatorClusterRole
	Namespace          string
	RegistryPullSecret string
	PSP                FlannelOperatorPSP
}

type FlannelOperatorClusterRole struct {
	BindingName string
	Name        string
}

type FlannelOperatorPSP struct {
	Name string
}

func NewFlannelOperator(config FlannelOperatorConfig) (string, error) {
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

	values, err := render.Render(flannelOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
