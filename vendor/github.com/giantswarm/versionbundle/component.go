package versionbundle

import (
	"encoding/json"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

// Component is the software component an authority provides. It describes the
// functionality of such a component being exposed by the authority. In return
// an authority guarantees to provide the components functionality.
type Component struct {
	// Name is the name of the exposed component.
	Name string `json:"name" yaml:"name"`
	// Version is the version of the exposed component.
	Version string `json:"version" yaml:"version"`
}

func (c Component) Validate() error {
	if c.Name == "" {
		return microerror.Maskf(invalidComponentError, "name must not be empty")
	}

	if c.Version == "" {
		return microerror.Maskf(invalidComponentError, "version must not be empty")
	}

	_, err := semver.NewVersion(c.Version)
	if err != nil {
		return microerror.Maskf(invalidComponentError, "version parsing failed with error %#q", err)
	}

	return nil
}

func CopyComponents(components []Component) []Component {
	raw, err := json.Marshal(components)
	if err != nil {
		panic(err)
	}

	var copy []Component
	err = json.Unmarshal(raw, &copy)
	if err != nil {
		panic(err)
	}

	return copy
}
