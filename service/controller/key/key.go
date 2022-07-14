package key

import (
	"github.com/giantswarm/aws-operator/v12/pkg/label"
	"github.com/giantswarm/aws-operator/v12/pkg/project"
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
