package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type APIExtensionsAWSConfigE2EConfig struct {
	CommonDomain         string
	ClusterName          string
	VersionBundleVersion string

	AWS APIExtensionsAWSConfigE2EConfigAWS
}

type APIExtensionsAWSConfigE2EConfigAWS struct {
	APIHostedZone     string
	IngressHostedZone string
	// NetworkCIDR is deprecated and optional meanwhile. When left empty IPAM is
	// activated. We still need defaults for older versions.
	NetworkCIDR string
	// PrivateSubnetCIDR is deprecated and optional meanwhile. When left empty
	// IPAM is activated. We still need defaults for older versions.
	PrivateSubnetCIDR string
	// PublicSubnetCIDR is deprecated and optional meanwhile. When left empty IPAM
	// is activated. We still need defaults for older versions.
	PublicSubnetCIDR string
	Region           string
	RouteTable0      string
	RouteTable1      string
	VPCPeerID        string
}

// NewAPIExtensionsAWSConfigE2E renders values required by apiextensions-aws-config-e2e-chart.
func NewAPIExtensionsAWSConfigE2E(config APIExtensionsAWSConfigE2EConfig) (string, error) {
	if config.CommonDomain == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}
	if config.ClusterName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterName must not be empty", config)
	}
	if config.VersionBundleVersion == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.VersionBundleVersion must not be empty", config)
	}
	if config.AWS.APIHostedZone == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS.APIHostedZone must not be empty", config)
	}
	if config.AWS.IngressHostedZone == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS.IngressHostedZone must not be empty", config)
	}
	if config.AWS.NetworkCIDR == "" {
		config.AWS.NetworkCIDR = "10.12.0.0/24"
	}
	if config.AWS.PrivateSubnetCIDR == "" {
		config.AWS.PrivateSubnetCIDR = "10.12.0.0/25"
	}
	if config.AWS.PublicSubnetCIDR == "" {
		config.AWS.PublicSubnetCIDR = "10.12.0.128/25"
	}
	if config.AWS.Region == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS.Region must not be empty", config)
	}
	if config.AWS.RouteTable0 == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS.RouteTable0 must not be empty", config)
	}
	if config.AWS.RouteTable1 == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS.RouteTable1 must not be empty", config)
	}
	if config.AWS.VPCPeerID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS.VPCPeerID must not be empty", config)
	}

	values, err := render.Render(apiExtensionsAWSConfigE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
