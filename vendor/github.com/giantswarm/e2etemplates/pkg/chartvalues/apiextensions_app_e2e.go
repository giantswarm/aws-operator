package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type APIExtensionsAppE2EConfig struct {
	App         APIExtensionsAppE2EConfigApp
	AppCatalog  APIExtensionsAppE2EConfigAppCatalog
	AppOperator APIExtensionsAppE2EConfigAppOperator
	ConfigMap   APIExtensionsAppE2EConfigConfigMap
	Namespace   string
	Secret      APIExtensionsAppE2EConfigSecret
}

type APIExtensionsAppE2EConfigApp struct {
	Config     APIExtensionsAppE2EConfigAppConfig
	Catalog    string
	KubeConfig APIExtensionsAppE2EConfigAppKubeConfig
	Name       string
	Namespace  string
	Version    string
}

type APIExtensionsAppE2EConfigAppCatalog struct {
	Name        string
	Title       string
	Description string
	LogoURL     string
	Storage     APIExtensionsAppE2EConfigAppCatalogStorage
}

type APIExtensionsAppE2EConfigAppCatalogStorage struct {
	Type string
	URL  string
}

type APIExtensionsAppE2EConfigAppConfig struct {
	ConfigMap APIExtensionsAppE2EConfigAppConfigConfigMap
	Secret    APIExtensionsAppE2EConfigAppConfigSecret
}

type APIExtensionsAppE2EConfigAppConfigConfigMap struct {
	Name      string
	Namespace string
}

type APIExtensionsAppE2EConfigAppConfigSecret struct {
	Name      string
	Namespace string
}

type APIExtensionsAppE2EConfigAppKubeConfig struct {
	InCluster bool
	Secret    APIExtensionsAppE2EConfigAppConfigKubeConfigSecret
}

type APIExtensionsAppE2EConfigAppConfigKubeConfigSecret struct {
	Name      string
	Namespace string
}

type APIExtensionsAppE2EConfigAppOperator struct {
	Version string
}

type APIExtensionsAppE2EConfigConfigMap struct {
	ValuesYAML string
}

type APIExtensionsAppE2EConfigSecret struct {
	ValuesYAML string
}

// NewAPIExtensionsAppE2E renders values required by
// apiextensions-app-config-e2e-chart.
func NewAPIExtensionsAppE2E(config APIExtensionsAppE2EConfig) (string, error) {
	if config.App.Catalog == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.App.Catalog must not be empty", config)
	}
	if config.App.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.App.Name must not be empty", config)
	}
	if config.App.Namespace == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.App.Namespace must not be empty", config)
	}
	if config.App.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.App.Version must not be empty", config)
	}
	if config.AppCatalog.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AppCatalog.Name must not be empty", config)
	}
	if config.AppCatalog.Storage.Type == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AppCatalog.Storage.Type must not be empty", config)
	}
	if config.AppCatalog.Storage.URL == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AppCatalog.Storage.URL must not be empty", config)
	}
	if config.AppOperator.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AppOperator.Version must not be empty", config)
	}
	if config.Namespace == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}

	values, err := render.Render(apiExtensionsAppE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
