package service

import (
	"github.com/giantswarm/versionbundle"

	clusterapiv29 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v29"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, clusterapiv29.VersionBundle())

	return versionBundles
}
