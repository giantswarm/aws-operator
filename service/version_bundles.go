package service

import (
	"github.com/giantswarm/versionbundle"

	v31 "github.com/giantswarm/aws-operator/service/controller/legacy/v31"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v31.VersionBundle())

	return versionBundles
}
