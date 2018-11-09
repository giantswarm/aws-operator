package service

import (
	"github.com/giantswarm/aws-operator/service/controller/v12"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1"
	"github.com/giantswarm/aws-operator/service/controller/v13"
	"github.com/giantswarm/aws-operator/service/controller/v14"
	"github.com/giantswarm/aws-operator/service/controller/v14patch1"
	"github.com/giantswarm/aws-operator/service/controller/v14patch2"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3"
	"github.com/giantswarm/aws-operator/service/controller/v15"
	"github.com/giantswarm/aws-operator/service/controller/v16"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1"
	"github.com/giantswarm/aws-operator/service/controller/v17"
	"github.com/giantswarm/aws-operator/service/controller/v18"
	"github.com/giantswarm/aws-operator/service/controller/v19"
	"github.com/giantswarm/versionbundle"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v12.VersionBundle())
	versionBundles = append(versionBundles, v12patch1.VersionBundle())
	versionBundles = append(versionBundles, v13.VersionBundle())
	versionBundles = append(versionBundles, v14.VersionBundle())
	versionBundles = append(versionBundles, v14patch1.VersionBundle())
	versionBundles = append(versionBundles, v14patch2.VersionBundle())
	versionBundles = append(versionBundles, v14patch3.VersionBundle())
	versionBundles = append(versionBundles, v15.VersionBundle())
	versionBundles = append(versionBundles, v16.VersionBundle())
	versionBundles = append(versionBundles, v16patch1.VersionBundle())
	versionBundles = append(versionBundles, v17.VersionBundle())
	versionBundles = append(versionBundles, v18.VersionBundle())
	versionBundles = append(versionBundles, v19.VersionBundle())

	return versionBundles
}
