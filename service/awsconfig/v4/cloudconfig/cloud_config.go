package cloudconfig

import (
	"strings"

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
	// Dependencies.
	Logger          micrologger.Logger
	K8sAPIExtraArgs []string
}

// DefaultConfig provides a default configuration to create a new cloud config
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:          nil,
		K8sAPIExtraArgs: []string{},
	}
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	// Dependencies.
	logger          micrologger.Logger
	k8sAPIExtraArgs []string
}

// New creates a new configured cloud config service.
func New(config Config) (*CloudConfig, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	var k8sAPIExtraArgs []string
	{
		for _, arg := range config.K8sAPIExtraArgs {
			if !strings.HasSuffix(arg, "=") {
				k8sAPIExtraArgs = append(k8sAPIExtraArgs, arg)
			}
		}
	}

	newCloudConfig := &CloudConfig{
		// Dependencies.
		logger:          config.Logger,
		k8sAPIExtraArgs: k8sAPIExtraArgs,
	}

	return newCloudConfig, nil
}
