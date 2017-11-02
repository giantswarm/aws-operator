package versionbundle

import "github.com/giantswarm/microerror"

type kind string

const (
	// KindAdded being used in a changelog describes an authority's component got
	// added.
	KindAdded kind = "added"
	// KindAdded being used in a changelog describes an authority's component got
	// changed.
	KindChanged kind = "changed"
	// KindAdded being used in a changelog describes an authority's component got
	// deprecated.
	KindDeprecated kind = "deprecated"
	// KindAdded being used in a changelog describes an authority's component got
	// fixed.
	KindFixed kind = "fixed"
	// KindAdded being used in a changelog describes an authority's component got
	// removed.
	KindRemoved kind = "removed"
	// KindAdded being used in a changelog describes an authority's component got
	// adapted for security reasons.
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
	Component string `json:"component" yaml:"component"`
	// Description is some text describing the changelog entry. This information
	// is intended to be useful for humans.
	Description string `json:"description" yaml:"description"`
	// Kind is a machine readable type describing what kind of changelog the
	// changelog actually is. Also see the kind type.
	Kind kind `json:"kind" yaml:"kind"`
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
