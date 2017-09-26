package cloudconfig

import (
	"github.com/giantswarm/certificatetpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	FileOwner      = "root:root"
	FilePermission = 0700
)

// Config represents the configuration used to create a cloud config service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloud config
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	// Dependencies.
	logger micrologger.Logger
}

// New creates a new configured cloud config service.
func New(config Config) (*CloudConfig, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	newCloudConfig := &CloudConfig{
		// Dependencies.
		logger: config.Logger,
	}

	return newCloudConfig, nil
}

// NewMasterTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewMasterTemplate(customObject kvmtpr.CustomObject, certs certificatetpr.AssetsBundle, node clustertprspec.Node) (string, error) {
	var err error

	// TODO remove defaulting as soon as custom objects are configured.
	if customObject.Spec.Cluster.Version == "" {
		customObject.Spec.Cluster.Version = string(cloudconfig.V_0_1_0)
	}

	var template string

	switch customObject.Spec.Cluster.Version {
	case string(cloudconfig.V_0_1_0):
		template, err = v_0_1_0MasterTemplate(customObject, certs, node)
		if err != nil {
			return "", microerror.Mask(err)
		}

	default:
		return "", microerror.Maskf(notFoundError, "k8scloudconfig version '%s'", customObject.Spec.Cluster.Version)
	}

	return template, nil
}

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(customObject kvmtpr.CustomObject, certs certificatetpr.AssetsBundle, node clustertprspec.Node) (string, error) {
	var err error

	// TODO remove defaulting as soon as custom objects are configured.
	if customObject.Spec.Cluster.Version == "" {
		customObject.Spec.Cluster.Version = string(cloudconfig.V_0_1_0)
	}

	var template string

	switch customObject.Spec.Cluster.Version {
	case string(cloudconfig.V_0_1_0):
		template, err = v_0_1_0WorkerTemplate(customObject, certs, node)
		if err != nil {
			return "", microerror.Mask(err)
		}

	default:
		return "", microerror.Maskf(notFoundError, "k8scloudconfig version '%s'", customObject.Spec.Cluster.Version)
	}

	return template, nil
}
