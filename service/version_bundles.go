package service

import (
	"github.com/giantswarm/versionbundle"

	clusterapiv31 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v31"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, clusterapiv31.VersionBundle())

	return versionBundles
}
