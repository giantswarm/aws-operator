package versionbundle

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	// OperatorEqual is the operator used to describe the mathematical comparison
	// '==' between matching dependency and component versions.
	OperatorEqual = "=="
	// OperatorGreater is the operator used to describe the mathematical
	// comparison '>' between matching dependency and component versions.
	OperatorGreater = ">"
	// OperatorGreaterOrEqual is the operator used to describe the mathematical
	// comparison '>=' between matching dependency and component versions.
	OperatorGreaterOrEqual = ">="
	// OperatorLess is the operator used to describe the mathematical comparison
	// '<' between matching dependency and component versions.
	OperatorLess = "<"
	// OperatorLessOrEqual is the operator used to describe the mathematical
	// comparison '<=' between matching dependency and component versions.
	OperatorLessOrEqual = "<="
	// OperatorNotEqual is the operator used to describe the mathematical
	// comparison '!=' between matching dependency and component versions.
	OperatorNotEqual = "!="
)

var (
	validOperators = []string{
		OperatorEqual,
		OperatorGreater,
		OperatorGreaterOrEqual,
		OperatorLess,
		OperatorLessOrEqual,
		OperatorNotEqual,
	}
)

// Dependency is the dependency an authority defines within its version bundles
// to describe the requirement of other components within the infrastructure to
// be able to reliably provide its own component suppurt.
type Dependency struct {
	// Name is the name of the exposed dependency.
	Name string `json:"name" yaml:"name"`
	// Version is the version of the exposed dependency. This version is a
	// combination of a supported mathematical operator and the actual semver
	// version separated by a space.
	//
	//     <operator> <major>.<minor>.<patch>
	//
	//     == 1.0.0
	//     != 1.5.0
	//     > 5.2.33
	//
	Version string `json:"version" yaml:"version"`
}

func (d Dependency) Matches(c Component) bool {
	dependencyVersion, componentVersion, operator := parseParameters(d.Version, c.Version)

	if operator == OperatorEqual {
		if componentVersion == dependencyVersion {
			return true
		}
	}

	if operator == OperatorGreaterOrEqual {
		if componentVersion >= dependencyVersion {
			return true
		}
	}

	if operator == OperatorGreater {
		if componentVersion > dependencyVersion {
			return true
		}
	}

	if operator == OperatorLess {
		if componentVersion < dependencyVersion {
			return true
		}
	}

	if operator == OperatorLessOrEqual {
		if componentVersion <= dependencyVersion {
			return true
		}
	}

	if operator == OperatorNotEqual {
		if componentVersion != dependencyVersion {
			return true
		}
	}

	return false
}

func (d Dependency) Validate() error {
	if d.Name == "" {
		return microerror.Maskf(invalidDependencyError, "name must not be empty")
	}

	if d.Version == "" {
		return microerror.Maskf(invalidDependencyError, "version must not be empty")
	}

	inputSplit := strings.Split(d.Version, " ")
	if len(inputSplit) != 2 {
		return microerror.Maskf(invalidDependencyError, "input format must be '<operator> <semver version>'")
	}

	operator := inputSplit[0]
	if operator == "" {
		return microerror.Maskf(invalidDependencyError, "operator must not be empty")
	}
	var found bool
	for _, o := range validOperators {
		if operator == o {
			found = true
		}
	}
	if !found {
		return microerror.Maskf(invalidDependencyError, "operator format must be one of %#v", validOperators)
	}

	versionSplit := strings.Split(inputSplit[1], ".")
	if len(versionSplit) != 3 {
		return microerror.Maskf(invalidDependencyError, "version format must be '<major>.<minor>.<patch>'")
	}

	if !isPositiveNumber(versionSplit[0]) {
		return microerror.Maskf(invalidDependencyError, "major version must be positive number")
	}

	minor := versionSplit[1]
	if !isPositiveNumber(minor) && minor != "x" {
		return microerror.Maskf(invalidDependencyError, "minor version must be positive number or wildcard ('x')")
	}

	patch := versionSplit[2]
	if !isPositiveNumber(patch) && patch != "x" {
		return microerror.Maskf(invalidDependencyError, "patch version must be positive number or wildcard ('x')")
	}

	if minor == "x" && patch != "x" {
		return microerror.Maskf(invalidDependencyError, "patch must be wildcard ('x') when minor is wildcard ('x')")
	}

	return nil
}

func parseParameters(dependencyVersion, componentVersion string) (string, string, string) {
	split := strings.Split(dependencyVersion, " ")
	dependencyVersion = split[1]

	i := strings.Index(dependencyVersion, "x")
	if i == -1 {
		return dependencyVersion, componentVersion, split[0]
	}

	return dependencyVersion[:i-1], componentVersion[:i-1], split[0]
}
