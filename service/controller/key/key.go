package key

import (
	"github.com/giantswarm/aws-operator/v15/pkg/label"
	"github.com/giantswarm/aws-operator/v15/pkg/project"
)

func VersionLabelSelector(enabled bool, overridenVersion string) string {
	if !enabled {
		return ""
	}

	version := project.Version()
	if overridenVersion != "" {
		version = overridenVersion
	}

	return label.OperatorVersion + "=" + version
}
