package service

import (
	"github.com/giantswarm/versionbundle"

	v12 "github.com/giantswarm/aws-operator/service/controller/v12"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1"
	v13 "github.com/giantswarm/aws-operator/service/controller/v13"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3"
	"github.com/giantswarm/aws-operator/service/controller/v14patch4"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1"
	v17 "github.com/giantswarm/aws-operator/service/controller/v17"
	"github.com/giantswarm/aws-operator/service/controller/v17patch1"
	v18 "github.com/giantswarm/aws-operator/service/controller/v18"
	v19 "github.com/giantswarm/aws-operator/service/controller/v19"
	v20 "github.com/giantswarm/aws-operator/service/controller/v20"
	v21 "github.com/giantswarm/aws-operator/service/controller/v21"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v12.VersionBundle())
	versionBundles = append(versionBundles, v12patch1.VersionBundle())
	versionBundles = append(versionBundles, v13.VersionBundle())
	versionBundles = append(versionBundles, v14patch3.VersionBundle())
	versionBundles = append(versionBundles, v14patch4.VersionBundle())
	versionBundles = append(versionBundles, v16patch1.VersionBundle())
	versionBundles = append(versionBundles, v17.VersionBundle())
	versionBundles = append(versionBundles, v17patch1.VersionBundle())
	versionBundles = append(versionBundles, v18.VersionBundle())
	versionBundles = append(versionBundles, v19.VersionBundle())
	versionBundles = append(versionBundles, v20.VersionBundle())
	versionBundles = append(versionBundles, v21.VersionBundle())

	return versionBundles
}
