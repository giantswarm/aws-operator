package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/aws-operator/service/controller/v1"
	"github.com/giantswarm/aws-operator/service/controller/v10"
	"github.com/giantswarm/aws-operator/service/controller/v11"
	"github.com/giantswarm/aws-operator/service/controller/v12"
	"github.com/giantswarm/aws-operator/service/controller/v2"
	"github.com/giantswarm/aws-operator/service/controller/v3"
	"github.com/giantswarm/aws-operator/service/controller/v4"
	"github.com/giantswarm/aws-operator/service/controller/v5"
	"github.com/giantswarm/aws-operator/service/controller/v6"
	"github.com/giantswarm/aws-operator/service/controller/v7"
	"github.com/giantswarm/aws-operator/service/controller/v8"
	"github.com/giantswarm/aws-operator/service/controller/v9"
	"github.com/giantswarm/aws-operator/service/controller/v9patch1"
	"github.com/giantswarm/aws-operator/service/controller/v9patch2"
)

// NewVersionBundles returns the array of version bundles defined for the
// operator.
func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v1.VersionBundles()...) // NOTE this is odd because of the version bundle introduction process.
	versionBundles = append(versionBundles, v2.VersionBundles()...) // NOTE this is odd because of the version bundle introduction process.
	versionBundles = append(versionBundles, v3.VersionBundles()...) // NOTE this is odd because of the version bundle introduction process.
	versionBundles = append(versionBundles, v4.VersionBundle())
	versionBundles = append(versionBundles, v5.VersionBundle())
	versionBundles = append(versionBundles, v6.VersionBundle())
	versionBundles = append(versionBundles, v7.VersionBundle())
	versionBundles = append(versionBundles, v8.VersionBundle())
	versionBundles = append(versionBundles, v9.VersionBundle())
	versionBundles = append(versionBundles, v9patch1.VersionBundle())
	versionBundles = append(versionBundles, v9patch2.VersionBundle())
	versionBundles = append(versionBundles, v10.VersionBundle())
	versionBundles = append(versionBundles, v11.VersionBundle())
	versionBundles = append(versionBundles, v12.VersionBundle())

	return versionBundles
}
