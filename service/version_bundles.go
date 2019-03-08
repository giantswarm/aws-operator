package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/aws-operator/service/controller/v22"
	"github.com/giantswarm/aws-operator/service/controller/v22patch1"
	"github.com/giantswarm/aws-operator/service/controller/v23"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1"
	"github.com/giantswarm/aws-operator/service/controller/v24"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v22.VersionBundle())
	versionBundles = append(versionBundles, v22patch1.VersionBundle())
	versionBundles = append(versionBundles, v23.VersionBundle())
	versionBundles = append(versionBundles, v23patch1.VersionBundle())
	versionBundles = append(versionBundles, v24.VersionBundle())

	return versionBundles
}
