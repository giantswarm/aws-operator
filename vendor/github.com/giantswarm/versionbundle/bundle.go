package versionbundle

import (
	"strings"

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
	// Name is the name of the authority exposing the version bundle.
	//
	// NOTE that once this property is set it must never change again.
	Name string `json:"name" yaml:"name"`
	// Provider describes infrastructure provider that is specific for this
	// Bundle.
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
	// Version describes the version of the version bundle. Versions of version
	// bundles must be semver versions. Versions must not be duplicated. Versions
	// should be incremented gradually.
	//
	// NOTE that once this property is set it must never change again.
	Version string `json:"version" yaml:"version"`
}

func (b Bundle) ID() string {
	n := strings.TrimSpace(b.Name)
	p := strings.TrimSpace(b.Provider)
	v := strings.TrimSpace(b.Version)
	return n + ":" + p + ":" + v
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

	for _, c := range b.Components {
		err := c.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	if b.Name == "" {
		return microerror.Maskf(invalidBundleError, "name must not be empty")
	}

	_, err := semver.NewVersion(b.Version)
	if err != nil {
		return microerror.Maskf(invalidBundleError, "version parsing failed with error %#q", err)
	}

	return nil
}
