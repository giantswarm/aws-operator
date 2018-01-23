package cloudconfigv1

import (
	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeytpr"
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
func (c *CloudConfig) NewMasterTemplate(customObject awstpr.CustomObject, certs certificatetpr.CompactTLSAssets, keys randomkeytpr.CompactRandomKeyAssets) (string, error) {
	var err error

	// Default the version if it is not configured or we are using Cloud Formation.
	// TODO Remove once Cloud Formation migration is complete.
	if !keyv1.HasClusterVersion(customObject) || keyv1.UseCloudFormation(customObject) {
		customObject.Spec.Cluster.Version = string("v_0_1_0")
	}

	var template string

	switch keyv1.ClusterVersion(customObject) {
	case string("v_0_1_0"):
		template, err = v_0_1_0MasterTemplate(customObject, certs, keys)
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
func (c *CloudConfig) NewWorkerTemplate(customObject awstpr.CustomObject, certs certificatetpr.CompactTLSAssets) (string, error) {
	var err error

	// Default the version if it is not configured.
	// TODO Remove once Cloud Formation migration is complete.
	if !keyv1.HasClusterVersion(customObject) {
		customObject.Spec.Cluster.Version = string("v_0_1_0")
	}

	var template string

	switch customObject.Spec.Cluster.Version {
	case string("v_0_1_0"):
		template, err = v_0_1_0WorkerTemplate(customObject, certs)
		if err != nil {
			return "", microerror.Mask(err)
		}

	default:
		return "", microerror.Maskf(notFoundError, "k8scloudconfig version '%s'", customObject.Spec.Cluster.Version)
	}

	return template, nil
}
