package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type APIExtensionsKVMConfigE2EConfig struct {
	ClusterID            string
	HttpNodePort         int
	HttpsNodePort        int
	VersionBundleVersion string
	VNI                  int
}

func NewAPIExtensionsKVMConfigE2E(config APIExtensionsKVMConfigE2EConfig) (string, error) {
	if config.ClusterID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.HttpNodePort < 0 {
		return "", microerror.Maskf(invalidConfigError, "%T.HttpNodePort must not be negative number", config)
	}
	if config.HttpsNodePort < 0 {
		return "", microerror.Maskf(invalidConfigError, "%T.HttpsNodePort must not be negative number", config)
	}
	if config.VersionBundleVersion == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.VersionBundleVersion must not be empty", config)
	}
	if config.VNI < 0 {
		return "", microerror.Maskf(invalidConfigError, "%T.VNI must not be negative number", config)
	}

	values, err := render.Render(apiExtensionsKVMConfigE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
