package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Modified to retrieve component versions from releases",
				Kind:        versionbundle.KindChanged,
			},
		},
		Name:    Name(),
		Version: BundleVersion(),
	}
}
