package service

import (
	"github.com/giantswarm/versionbundle"

	clusterapiv30 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v30"
	v25 "github.com/giantswarm/aws-operator/service/controller/legacy/v25"
	v26 "github.com/giantswarm/aws-operator/service/controller/legacy/v26"
	v27 "github.com/giantswarm/aws-operator/service/controller/legacy/v27"
	v28 "github.com/giantswarm/aws-operator/service/controller/legacy/v28"
	v28patch1 "github.com/giantswarm/aws-operator/service/controller/legacy/v28patch1"
	v29 "github.com/giantswarm/aws-operator/service/controller/legacy/v29"
	v29patch1 "github.com/giantswarm/aws-operator/service/controller/legacy/v29patch1"
	v30 "github.com/giantswarm/aws-operator/service/controller/legacy/v30"
	v31 "github.com/giantswarm/aws-operator/service/controller/legacy/v31"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, clusterapiv30.VersionBundle())

	versionBundles = append(versionBundles, v25.VersionBundle())
	versionBundles = append(versionBundles, v26.VersionBundle())
	versionBundles = append(versionBundles, v27.VersionBundle())
	versionBundles = append(versionBundles, v28.VersionBundle())
	versionBundles = append(versionBundles, v28patch1.VersionBundle())
	versionBundles = append(versionBundles, v29.VersionBundle())
	versionBundles = append(versionBundles, v29patch1.VersionBundle())
	versionBundles = append(versionBundles, v30.VersionBundle())
	versionBundles = append(versionBundles, v31.VersionBundle())

	return versionBundles
}
