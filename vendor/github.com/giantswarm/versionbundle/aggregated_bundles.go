package versionbundle

import (
	"reflect"

	"github.com/giantswarm/microerror"
)

// AggregatedBundles is a plain validation type for aggregated lists of version
// bundles. Lists of version bundles reflect releases.
type AggregatedBundles [][]Bundle

func (b AggregatedBundles) Validate() error {
	if len(b) != 0 {
		l := len(b[0])
		for _, group := range b {
			if l != len(group) {
				return microerror.Maskf(invalidAggregatedBundlesError, "number of version bundles within aggregated version bundles must be equal")
			}
		}
	}

	if b.hasDuplicatedAggregatedBundles() {
		return microerror.Maskf(invalidAggregatedBundlesError, "version bundles within aggregated version bundles must be unique")
	}

	return nil
}

func (b AggregatedBundles) hasDuplicatedAggregatedBundles() bool {
	for _, b1 := range b {
		var seen int

		for _, b2 := range b {
			if reflect.DeepEqual(b1, b2) {
				seen++

				if seen >= 2 {
					return true
				}
			}
		}
	}

	return false
}
