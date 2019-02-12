package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type APIExtensionsChartE2EConfig struct {
	Chart         APIExtensionsChartE2EConfigChart
	ChartOperator APIExtensionsChartE2EConfigChartOperator
	ConfigMap     APIExtensionsChartE2EConfigConfigMap
	Namespace     string
	Secret        APIExtensionsChartE2EConfigSecret
}

type APIExtensionsChartE2EConfigChart struct {
	Config     APIExtensionsChartE2EConfigChartConfig
	Name       string
	Namespace  string
	TarballURL string
}

type APIExtensionsChartE2EConfigChartConfig struct {
	ConfigMap APIExtensionsChartE2EConfigChartConfigConfigMap
	Secret    APIExtensionsChartE2EConfigChartConfigSecret
}

type APIExtensionsChartE2EConfigChartConfigConfigMap struct {
	Name      string
	Namespace string
}

type APIExtensionsChartE2EConfigChartConfigSecret struct {
	Name      string
	Namespace string
}

type APIExtensionsChartE2EConfigChartOperator struct {
	Version string
}

type APIExtensionsChartE2EConfigConfigMap struct {
	ValuesYAML string
}

type APIExtensionsChartE2EConfigSecret struct {
	ValuesYAML string
}

// NewAPIExtensionsChartE2E renders values required by
// apiextensions-azure-config-e2e-chart.
func NewAPIExtensionsChartE2E(config APIExtensionsChartE2EConfig) (string, error) {
	if config.Chart.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Chart.Name must not be empty", config)
	}
	if config.Chart.Namespace == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Chart.Namespace must not be empty", config)
	}
	if config.Chart.TarballURL == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Chart.TarballURL must not be empty", config)
	}
	if config.ChartOperator.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ChartOperator.Version must not be empty", config)
	}
	if config.Namespace == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}

	values, err := render.Render(apiExtensionsChartE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
