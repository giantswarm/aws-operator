package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Name:    Name(),
		Version: BundleVersion(),
	}
}
