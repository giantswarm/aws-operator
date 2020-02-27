package versionbundle

import (
	"encoding/json"
	"strings"

	"github.com/giantswarm/microerror"
)

type kind string

const (
	// KindAdded is used in changelogs for new features.
	KindAdded kind = "added"
	// KindChanged is used in changelogs for changes in existing functionality.
	KindChanged kind = "changed"
	// KindDeprecated is used in a changelogs for soon-to-be removed features.
	KindDeprecated kind = "deprecated"
	// KindFixed is used in changelogs for any bug fixes.
	KindFixed kind = "fixed"
	// KindRemoved is used in changelogs for now removed features.
	KindRemoved kind = "removed"
	// KindSecurity is used in chnagelogs in case of vulnerabilities.
	KindSecurity kind = "security"
)

var (
	validKinds = []kind{
		KindAdded,
		KindChanged,
		KindDeprecated,
		KindFixed,
		KindRemoved,
		KindSecurity,
	}
)

// Changelog is a single changelog entry a version bundle must define. Its
// intention is to explain the introduction of the version bundle.
type Changelog struct {
	// Component is the component the changelog is about. Thus might be a
	// component provided by another authority. To be able to properly aggregate
	// version bundles the given component must exist, either within the same
	// authority or within another authority within the infrastructure. That is,
	// Aggregate must know about it to be able to properly merge version bundles.
	Component string `json:"component,omitempty" yaml:"component,omitempty"`
	// ComponentVersion is the upstream version of the exposed component.
	ComponentVersion string `json:"componentVersion,omitempty" yaml:"componentVersion,omitempty"`
	// Description is some text describing the changelog entry. This information
	// is intended to be useful for humans.
	Description string `json:"description" yaml:"description"`
	// Kind is a machine readable type describing what kind of changelog the
	// changelog actually is. Also see the kind type.
	Kind kind `json:"kind" yaml:"kind"`
	// URLs is a list of links which contain additional information to the
	// changelog entry such as upstream changelogs or pull requests.
	URLs []string `json:"urls" yaml:"urls"`
	// Version is the Giant Swarm version of the component.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

func (c Changelog) String() string {
	return string(c.Kind) + ":" + c.Component + ":" + c.Description
}

func (c Changelog) Validate() error {
	if c.Component == "" {
		return microerror.Maskf(invalidChangelogError, "component must not be empty")
	}

	if c.Description == "" {
		return microerror.Maskf(invalidChangelogError, "description must not be empty")
	}

	if c.Kind == "" {
		return microerror.Maskf(invalidChangelogError, "kind must not be empty")
	}

	var found bool
	for _, k := range validKinds {
		if c.Kind == k {
			found = true
		}
	}
	if !found {
		return microerror.Maskf(invalidChangelogError, "kind must be one of %#v", validKinds)
	}

	return nil
}

func CopyChangelogs(changelogs []Changelog) []Changelog {
	raw, err := json.Marshal(changelogs)
	if err != nil {
		panic(err)
	}

	var copy []Changelog
	err = json.Unmarshal(raw, &copy)
	if err != nil {
		panic(err)
	}

	return copy
}

func NewKind(kindType string) (kind, error) {
	converted := kind(strings.ToLower(kindType))
	for _, k := range validKinds {
		if converted == k {
			return converted, nil
		}
	}
	return kind(""), microerror.Maskf(executionFailedError, "kind must be one of %#v", validKinds)
}
