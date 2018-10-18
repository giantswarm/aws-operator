package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type APIExtensionsAzureConfigE2EConfig struct {
	Azure                     APIExtensionsAzureConfigE2EConfigAzure
	ClusterName               string
	CommonDomain              string
	CommonDomainResourceGroup string
	VersionBundleVersion      string
}

type APIExtensionsAzureConfigE2EConfigAzure struct {
	CalicoSubnetCIDR string
	CIDR             string
	Location         string
	MasterSubnetCIDR string
	VMSizeMaster     string
	VMSizeWorker     string
	VPNSubnetCIDR    string
	WorkerSubnetCIDR string
}

// NewAPIExtensionsAzureConfigE2E renders values required by
// apiextensions-azure-config-e2e-chart.
func NewAPIExtensionsAzureConfigE2E(config APIExtensionsAzureConfigE2EConfig) (string, error) {
	if config.Azure.CalicoSubnetCIDR == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Azure.CalicoSubnetCIDR must not be empty", config)
	}
	if config.Azure.CIDR == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Azure.CIDR must not be empty", config)
	}
	if config.Azure.Location == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Azure.Location must not be empty", config)
	}
	if config.Azure.MasterSubnetCIDR == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Azure.MasterSubnetCIDR must not be empty", config)
	}
	if config.Azure.VMSizeMaster == "" {
		config.Azure.VMSizeMaster = "Standard_D2s_v3"
	}
	if config.Azure.VMSizeWorker == "" {
		config.Azure.VMSizeWorker = "Standard_D2s_v3"
	}
	if config.Azure.VPNSubnetCIDR == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Azure.VPNSubnetCIDR must not be empty", config)
	}
	if config.Azure.WorkerSubnetCIDR == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Azure.WorkerSubnetCIDR must not be empty", config)
	}
	if config.ClusterName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterName must not be empty", config)
	}
	if config.CommonDomain == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}
	if config.CommonDomainResourceGroup == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.CommonDomainResourceGroup must not be empty", config)
	}
	if config.VersionBundleVersion == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.VersionBundleVersion must not be empty", config)
	}

	values, err := render.Render(apiExtensionsAzureConfigE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
