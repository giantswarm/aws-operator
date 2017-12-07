package cloudconfigv2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeytpr"

	"github.com/giantswarm/aws-operator/service/keyv2"
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
func (c *CloudConfig) NewMasterTemplate(customObject v1alpha1.AWSConfig, certs certificatetpr.CompactTLSAssets, keys randomkeytpr.CompactRandomKeyAssets) (string, error) {
	var err error

	// Default the version if it is not configured or we are using Cloud Formation.
	// TODO Remove once Cloud Formation migration is complete.
	if !keyv2.HasClusterVersion(customObject) || keyv2.UseCloudFormation(customObject) {
		customObject.Spec.Cluster.Version = string(cloudconfig.V_0_1_0)
	}

	var template string

	switch keyv2.ClusterVersion(customObject) {
	case string(cloudconfig.V_0_1_0):
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
func (c *CloudConfig) NewWorkerTemplate(customObject v1alpha1.AWSConfig, certs certificatetpr.CompactTLSAssets) (string, error) {
	var err error

	// Default the version if it is not configured.
	// TODO Remove once Cloud Formation migration is complete.
	if !keyv2.HasClusterVersion(customObject) {
		customObject.Spec.Cluster.Version = string(cloudconfig.V_0_1_0)
	}

	var template string

	switch customObject.Spec.Cluster.Version {
	case string(cloudconfig.V_0_1_0):
		template, err = v_0_1_0WorkerTemplate(customObject, certs)
		if err != nil {
			return "", microerror.Mask(err)
		}

	default:
		return "", microerror.Maskf(notFoundError, "k8scloudconfig version '%s'", customObject.Spec.Cluster.Version)
	}

	return template, nil
}
