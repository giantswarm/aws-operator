package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/aws-operator/service/key"
)

const (
	legacyResourceName = "legacy"
)

// NewResourceRouter determines which resources are enabled based upon the
// version in the version bundle.
func NewResourceRouter(resources map[string][]framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		var enabledResources []framework.Resource

		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return enabledResources, microerror.Mask(err)
		}

		switch key.VersionBundleVersion(customObject) {
		case key.LegacyVersion:
			// Legacy version so only enable the legacy resource.
			enabledResources = resources[key.LegacyVersion]
		case key.CloudFormationVersion:
			// Cloud Formation transitional version so enable all resources.
			enabledResources = resources[key.CloudFormationVersion]
		case "":
			// Default to the legacy resource for custom objects without a version
			// bundle.
			enabledResources = resources[key.LegacyVersion]
		default:
			return enabledResources, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", key.VersionBundleVersion(customObject))
		}

		return enabledResources, nil
	}
}
