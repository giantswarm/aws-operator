package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/aws-operator/service/awsconfig/v1"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2"
	"github.com/giantswarm/aws-operator/service/awsconfig/v3"
	"github.com/giantswarm/aws-operator/service/awsconfig/v4"
	"github.com/giantswarm/aws-operator/service/awsconfig/v5"
	"github.com/giantswarm/aws-operator/service/awsconfig/v6"
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

	return versionBundles
}
