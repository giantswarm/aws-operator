package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	FileOwner          = "root:root"
	FilePermission     = 0700
	GzipBase64Encoding = "gzip+base64"
)

// Config represents the configuration used to create a cloud config service.
type Config struct {
	KMSClient KMSClient
	Logger    micrologger.Logger

	OIDC OIDCConfig
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	kmsClient KMSClient
	logger    micrologger.Logger

	k8sAPIExtraArgs []string
}

// OIDCConfig represents the configuration of the OIDC authorization provider
type OIDCConfig struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

// New creates a new configured cloud config service.
func New(config Config) (*CloudConfig, error) {
	if config.KMSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.KMSClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var k8sAPIExtraArgs []string
	{
		if config.OIDC.ClientID != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-client-id=%s", config.OIDC.ClientID))
		}
		if config.OIDC.IssuerURL != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", config.OIDC.IssuerURL))
		}
		if config.OIDC.UsernameClaim != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", config.OIDC.UsernameClaim))
		}
		if config.OIDC.GroupsClaim != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", config.OIDC.GroupsClaim))
		}
	}

	newCloudConfig := &CloudConfig{
		kmsClient: config.KMSClient,
		logger:    config.Logger,

		k8sAPIExtraArgs: k8sAPIExtraArgs,
	}

	return newCloudConfig, nil
}
