package key

import (
	"github.com/giantswarm/aws-operator/pkg/project"
)

func VersionLabelSelector(enable bool, overrideVersion string) string {
	if !enable {
		return ""
	}

	version := project.Version()
	if overrideVersion != "" {
		version = overrideVersion
	}

	return project.Name() + ".giantswarm.io/version=" + version
}
