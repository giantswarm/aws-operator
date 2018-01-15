package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/aws-operator/service/keyv2"
)

const (
	legacyResourceName = "legacy"
)

// NewResourceRouter determines which resources are enabled based upon the
// version in the version bundle.
func NewResourceRouter(resources map[string][]framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		var enabledResources []framework.Resource

		customObject, err := keyv2.ToCustomObject(obj)
		if err != nil {
			return enabledResources, microerror.Mask(err)
		}

		version := keyv2.VersionBundleVersion(customObject)

		switch version {
		case "0.1.0":
			// Legacy version so only enable the legacy resource.
			enabledResources = resources[version]
		case "0.2.0":
			// Cloud Formation transitional version so enable all resources.
			enabledResources = resources[version]
		case "1.0.0":
			// Kubernetes update to 1.8.4.
			enabledResources = resources[version]
		case "2.0.0":
			// First release of Cloud Formation based operator.
			enabledResources = resources[version]
		case "":
			// Default to the legacy resource for custom objects without a version
			// bundle.
			enabledResources = resources[keyv2.LegacyVersion]
		default:
			return enabledResources, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", keyv2.VersionBundleVersion(customObject))
		}

		return enabledResources, nil
	}
}

// NewTPRResourceRouter determines which resources are enabled based upon the
// version in the version bundle. It will be removed when we remove TPR support.
func NewTPRResourceRouter(resources map[string][]framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		var enabledResources []framework.Resource

		customObject, err := keyv1.ToCustomObject(obj)
		if err != nil {
			return enabledResources, microerror.Mask(err)
		}

		version := keyv1.VersionBundleVersion(customObject)

		switch version {
		case "0.1.0":
			// Legacy version so only enable the legacy resource.
			enabledResources = resources[version]
		case "0.2.0":
			// Cloud Formation transitional version so enable all resources.
			enabledResources = resources[version]
		case "1.0.0":
			// Kubernetes update to 1.8.4.
			enabledResources = resources[version]
		case "2.0.0":
			// First release of Cloud Formation based operator.
			enabledResources = resources[version]
		case "":
			// Default to the legacy resource for custom objects without a version
			// bundle.
			enabledResources = resources[keyv1.LegacyVersion]
		default:
			return enabledResources, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", keyv1.VersionBundleVersion(customObject))
		}

		return enabledResources, nil
	}
}
