package key

import (
	"fmt"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
)

func VersionLabelSelector(enabled bool, overridenVersion string) string {
	s := pawelVersionLabelSelector(enabled, overridenVersion)

	fmt.Printf("pawel ******* enabled=%#v overridenVersion=%#v out=%#v\n", enabled, overridenVersion, s)

	return s
}

func pawelVersionLabelSelector(enabled bool, overridenVersion string) string {
	if !enabled {
		return ""
	}

	version := project.Version()
	if overridenVersion != "" {
		version = overridenVersion
	}

	return label.OperatorVersion + "=" + version
}
