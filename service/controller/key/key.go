package key

import (
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
)

func VersionLabelSelector(enabled bool, overridedVersion string) string {
	if !enabled {
		return ""
	}

	version := project.Version()
	if overridedVersion != "" {
		version = overridedVersion
	}

	return label.OperatorVersion + "=" + version
}
