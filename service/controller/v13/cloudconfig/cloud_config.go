package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
)

const (
	FileOwner          = "root:root"
	FilePermission     = 0700
	GzipBase64Encoding = "gzip+base64"
)

// Config represents the configuration used to create a cloud config service.
type Config struct {
	Encrypter encrypter.Interface
	Logger    micrologger.Logger

	OIDC                   OIDCConfig
	PodInfraContainerImage string
	RegistryDomain         string
	SSOPublicKey           string
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	encrypter encrypter.Interface
	logger    micrologger.Logger

	k8sAPIExtraArgs     []string
	k8sKubeletExtraArgs []string
	registryDomain      string
	SSOPublicKey        string
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
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
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

	var k8sKubeletExtraArgs []string
	{
		if config.PodInfraContainerImage != "" {
			k8sKubeletExtraArgs = append(k8sKubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", config.PodInfraContainerImage))
		}
	}

	newCloudConfig := &CloudConfig{
		encrypter: config.Encrypter,
		logger:    config.Logger,

		k8sAPIExtraArgs:     k8sAPIExtraArgs,
		k8sKubeletExtraArgs: k8sKubeletExtraArgs,
		registryDomain:      config.RegistryDomain,
		SSOPublicKey:        config.SSOPublicKey,
	}

	return newCloudConfig, nil
}
