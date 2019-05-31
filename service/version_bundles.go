package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v22"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v22patch1"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v23"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v24"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v25"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v27"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v28"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v22.VersionBundle())
	versionBundles = append(versionBundles, v22patch1.VersionBundle())
	versionBundles = append(versionBundles, v23.VersionBundle())
	versionBundles = append(versionBundles, v24.VersionBundle())
	versionBundles = append(versionBundles, v25.VersionBundle())
	versionBundles = append(versionBundles, v26.VersionBundle())
	versionBundles = append(versionBundles, v27.VersionBundle())
	versionBundles = append(versionBundles, v28.VersionBundle())

	return versionBundles
}
