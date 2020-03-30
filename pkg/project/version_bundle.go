package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{},
		Name:       Name(),
		Version:    BundleVersion(),
	}
}
