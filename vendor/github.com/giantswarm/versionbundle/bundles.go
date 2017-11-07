package versionbundle

import (
	"encoding/json"
	"reflect"
	"sort"

	"github.com/giantswarm/microerror"
)

// Bundles is a plain validation type for a list of version bundles. A
// list of version bundles is exposed by authorities. Lists of version bundles
// of multiple authorities are aggregated and grouped to reflect distributions.
type Bundles []Bundle

func (b Bundles) Contain(item Bundle) bool {
	for _, bundle := range b {
		if reflect.DeepEqual(bundle, item) {
			return true
		}
	}

	return false
}

func (b Bundles) Validate() error {
	if len(b) == 0 {
		return microerror.Maskf(invalidBundlesError, "version bundles must not be empty")
	}

	if b.hasDuplicatedVersions() {
		return microerror.Maskf(invalidBundlesError, "version bundle versions must be unique")
	}

	b1 := CopyBundles(b)
	b2 := CopyBundles(b)
	sort.Sort(SortBundlesByTime(b1))
	sort.Sort(SortBundlesByVersion(b2))
	if !reflect.DeepEqual(b1, b2) {
		return microerror.Maskf(invalidBundlesError, "version bundle versions must always increment")
	}

	for _, bundle := range b {
		err := bundle.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundlesError, err.Error())
		}
	}

	bundleName := b[0].Name
	for _, bundle := range b {
		if bundle.Name != bundleName {
			return microerror.Maskf(invalidBundlesError, "name must be the same for all version bundles")
		}
	}

	return nil
}

func CopyBundles(bundles []Bundle) []Bundle {
	raw, err := json.Marshal(bundles)
	if err != nil {
		panic(err)
	}

	var copy []Bundle
	err = json.Unmarshal(raw, &copy)
	if err != nil {
		panic(err)
	}

	return copy
}

func (b Bundles) hasDuplicatedVersions() bool {
	for _, b1 := range b {
		var seen int

		for _, b2 := range b {
			if b1.Version == b2.Version {
				seen++

				if seen >= 2 {
					return true
				}
			}
		}
	}

	return false
}
