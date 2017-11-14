package versionbundle

import (
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

// Bundle represents a single version bundle exposed by an authority. An
// authority might exposes mutliple version bundles using the Capability
// structure. Version bundles are aggregated into a merged structure represented
// by the Aggregation structure. Also see the Aggregate function.
type Bundle struct {
	// Changelogs describe what changes are introduced by the version bundle. Each
	// version bundle must have at least one changelog entry.
	//
	// NOTE that once this property is set it must never change again.
	Changelogs []Changelog `json:"changelogs" yaml:"changelogs"`
	// Components describe the components an authority exposes. Functionality of
	// components listed here is guaranteed to be implemented in the according
	// versions.
	//
	// NOTE that once this property is set it must never change again.
	Components []Component `json:"components" yaml:"components"`
	// Dependencies describe which components other authorities expose have to be
	// available to be able to guarantee functionality this authority implements.
	//
	// NOTE that once this property is set it must never change again.
	Dependencies []Dependency `json:"dependencies" yaml:"dependencies"`
	// Deprecated defines a version bundle to be deprecated. Deprecated version
	// bundles are not intended to be mainatined anymore. Further usage of a
	// deprecated version bundle should be omitted.
	Deprecated bool `json:"deprecated" yaml:"deprecated"`
	// Name is the name of the authority exposing the version bundle.
	//
	// NOTE that once this property is set it must never change again.
	Name string `json:"name" yaml:"name"`
	// Time describes the time this version bundle got introduced.
	//
	// NOTE that once this property is set it must never change again.
	Time time.Time `json:"time" yaml:"time"`
	// Version describes the version of the version bundle. Versions of version
	// bundles must be semver versions. Versions must not be duplicated. Versions
	// should be incremented gradually.
	//
	// NOTE that once this property is set it must never change again.
	Version string `json:"version" yaml:"version"`
	// WIP describes if a version bundle is being developed. Usage of a version
	// bundle still being developed should be omitted.
	WIP bool `json:"wip" yaml:"wip"`
}

func (b Bundle) IsMajorUpgrade(other Bundle) (bool, error) {
	err := b.Validate()
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}
	err = other.Validate()
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}

	if b.Name != other.Name {
		return false, microerror.Maskf(invalidBundleError, "bundle must be from the same authority")
	}

	bSemver := semver.New(b.Version)
	otherSemver := semver.New(other.Version)
	if bSemver.Compare(*otherSemver) >= 0 {
		return false, nil
	}

	if bSemver.Major < otherSemver.Major {
		return true, nil
	}

	return false, nil
}

func (b Bundle) IsMinorUpgrade(other Bundle) (bool, error) {
	err := b.Validate()
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}
	err = other.Validate()
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}

	if b.Name != other.Name {
		return false, microerror.Maskf(invalidBundleError, "bundle must be from the same authority")
	}

	bSemver := semver.New(b.Version)
	otherSemver := semver.New(other.Version)
	if bSemver.Compare(*otherSemver) >= 0 {
		return false, nil
	}

	isMajor, err := b.IsMajorUpgrade(other)
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}
	if isMajor {
		return false, nil
	}

	if bSemver.Minor < otherSemver.Minor {
		return true, nil
	}

	return false, nil
}

func (b Bundle) IsPatchUpgrade(other Bundle) (bool, error) {
	err := b.Validate()
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}
	err = other.Validate()
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}

	if b.Name != other.Name {
		return false, microerror.Maskf(invalidBundleError, "bundle must be from the same authority")
	}

	bSemver := semver.New(b.Version)
	otherSemver := semver.New(other.Version)
	if bSemver.Compare(*otherSemver) >= 0 {
		return false, nil
	}

	isMajor, err := b.IsMajorUpgrade(other)
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}
	if isMajor {
		return false, nil
	}

	isMinor, err := b.IsMinorUpgrade(other)
	if err != nil {
		return false, microerror.Maskf(invalidBundleError, err.Error())
	}
	if isMinor {
		return false, nil
	}

	if bSemver.Patch < otherSemver.Patch {
		return true, nil
	}

	return false, nil
}

func (b Bundle) Validate() error {
	if len(b.Changelogs) == 0 {
		return microerror.Maskf(invalidBundleError, "changelogs must not be empty")
	}
	for _, c := range b.Changelogs {
		err := c.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	if len(b.Components) == 0 {
		return microerror.Maskf(invalidBundleError, "components must not be empty")
	}
	for _, c := range b.Components {
		err := c.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	for _, d := range b.Dependencies {
		err := d.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	var emptyTime time.Time
	if b.Time == emptyTime {
		return microerror.Maskf(invalidBundleError, "time must not be empty")
	}

	if b.Name == "" {
		return microerror.Maskf(invalidBundleError, "name must not be empty")
	}

	versionSplit := strings.Split(b.Version, ".")
	if len(versionSplit) != 3 {
		return microerror.Maskf(invalidBundleError, "version format must be '<major>.<minor>.<patch>'")
	}

	if !isPositiveNumber(versionSplit[0]) {
		return microerror.Maskf(invalidBundleError, "major version must be positive number")
	}

	if !isPositiveNumber(versionSplit[1]) {
		return microerror.Maskf(invalidBundleError, "minor version must be positive number")
	}

	if !isPositiveNumber(versionSplit[2]) {
		return microerror.Maskf(invalidBundleError, "patch version must be positive number")
	}

	return nil
}
