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
func NewResourceRouter(rs []framework.Resource) func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
	return func(ctx context.Context, obj interface{}) ([]framework.Resource, error) {
		var enabledResources []framework.Resource

		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return enabledResources, microerror.Mask(err)
		}

		switch key.VersionBundleVersion(customObject) {
		case key.LegacyVersion:
			// Legacy version so only enable the legacy resource.
			enabledResources = filterResourcesByName(rs, legacyResourceName)
		case key.CloudFormationVersion:
			// Cloud Formation transitional version so enable all resources.
			enabledResources = rs
		default:
			// Default to the legacy resource for custom objects without a version
			// bundle. TODO Return an error once the legacy resource is deprecated.
			enabledResources = filterResourcesByName(rs, legacyResourceName)
		}

		return enabledResources, nil
	}
}

// filterResourcesByName filters a list of resources by one or more resource
// names. It checks both the resource name and the underlying resource name
// in case the resources have been wrapped.
func filterResourcesByName(resources []framework.Resource, resourceNames ...string) []framework.Resource {
	resourceLookup := make(map[string]bool)

	for _, resourceName := range resourceNames {
		resourceLookup[resourceName] = true
	}

	var enabledResources []framework.Resource

	for _, resource := range resources {
		if _, ok := resourceLookup[resource.Name()]; ok {
			enabledResources = append(enabledResources, resource)
			continue
		}

		if _, ok := resourceLookup[resource.Underlying().Name()]; ok {
			enabledResources = append(enabledResources, resource)
		}
	}

	return enabledResources
}
