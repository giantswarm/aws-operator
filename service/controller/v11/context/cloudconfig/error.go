package cloudconfig

import "github.com/giantswarm/microerror"

var cloudConfigNotFound = microerror.New("cloud config not found")

// IsCloudConfigNotFoundError asserts cloudConfigNotFound.
func IsCloudConfigNotFoundError(err error) bool {
	return microerror.Cause(err) == cloudConfigNotFound
}
