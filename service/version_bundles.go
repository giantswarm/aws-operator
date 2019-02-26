package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/aws-operator/service/controller/v17patch1"
	"github.com/giantswarm/aws-operator/service/controller/v17patch2"
	"github.com/giantswarm/aws-operator/service/controller/v18"
	"github.com/giantswarm/aws-operator/service/controller/v19"
	"github.com/giantswarm/aws-operator/service/controller/v20"
	"github.com/giantswarm/aws-operator/service/controller/v21"
	"github.com/giantswarm/aws-operator/service/controller/v21patch1"
	"github.com/giantswarm/aws-operator/service/controller/v22"
	"github.com/giantswarm/aws-operator/service/controller/v23"
	"github.com/giantswarm/aws-operator/service/controller/v24"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v17patch1.VersionBundle())
	versionBundles = append(versionBundles, v17patch2.VersionBundle())
	versionBundles = append(versionBundles, v18.VersionBundle())
	versionBundles = append(versionBundles, v19.VersionBundle())
	versionBundles = append(versionBundles, v20.VersionBundle())
	versionBundles = append(versionBundles, v21.VersionBundle())
	versionBundles = append(versionBundles, v21patch1.VersionBundle())
	versionBundles = append(versionBundles, v22.VersionBundle())
	versionBundles = append(versionBundles, v23.VersionBundle())
	versionBundles = append(versionBundles, v24.VersionBundle())

	return versionBundles
}
