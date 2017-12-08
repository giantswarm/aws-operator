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

		switch keyv2.VersionBundleVersion(customObject) {
		case keyv2.LegacyVersion:
			// Legacy version so only enable the legacy resource.
			enabledResources = resources[keyv2.LegacyVersion]
		case keyv2.CloudFormationVersion:
			// Cloud Formation transitional version so enable all resources.
			enabledResources = resources[keyv2.CloudFormationVersion]
		case "1.0.0":
			// Kubernetes update to 1.8.4.
			enabledResources = resources["1.0.0"]
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

		switch keyv1.VersionBundleVersion(customObject) {
		case keyv1.LegacyVersion:
			// Legacy version so only enable the legacy resource.
			enabledResources = resources[keyv1.LegacyVersion]
		case keyv1.CloudFormationVersion:
			// Cloud Formation transitional version so enable all resources.
			enabledResources = resources[keyv1.CloudFormationVersion]
		case "1.0.0":
			// Kubernetes update to 1.8.4.
			enabledResources = resources["1.0.0"]
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
