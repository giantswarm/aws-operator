package awsconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
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

		version := key.VersionBundleVersion(customObject)
		enabledResources, ok := resources[version]
		if !ok {
			return enabledResources, microerror.Maskf(invalidVersionError, "version '%s' in version bundle is invalid", version)
		}

		return enabledResources, nil
	}
}
