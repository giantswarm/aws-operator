package versionbundle

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type AggregatorConfig struct {
	Logger micrologger.Logger
}

type Aggregator struct {
	logger micrologger.Logger
}

func NewAggregator(config AggregatorConfig) (*Aggregator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &Aggregator{
		logger: config.Logger,
	}

	return a, nil
}

// Aggregate merges version bundles based on dependencies each version bundle
// within the given version bundles define for their own components.
func (a *Aggregator) Aggregate(bundles []Bundle) ([][]Bundle, error) {
	if len(bundles) == 0 {
		return nil, nil
	}

	var aggregatedBundles [][]Bundle

	if len(bundles) == 1 {
		aggregatedBundles = append(aggregatedBundles, bundles)
		return aggregatedBundles, nil
	}

	for _, b1 := range bundles {
		newGroup := []Bundle{
			b1,
		}

		for _, b2 := range bundles {
			if reflect.DeepEqual(b1, b2) {
				continue
			}

			if a.bundlesConflictWithDependencies(b1, b2) {
				continue
			}

			if a.bundlesConflictWithDependencies(b2, b1) {
				continue
			}

			if a.containsBundleByName(newGroup, b2) {
				continue
			}

			newGroup = append(newGroup, b2)
		}

		sort.Sort(SortBundlesByVersion(newGroup))
		sort.Stable(SortBundlesByName(newGroup))

		if a.containsAggregatedBundle(aggregatedBundles, newGroup) {
			continue
		}

		if a.aggregatedBundlesMissVersionBundle(bundles, newGroup) {
			continue
		}

		aggregatedBundles = append(aggregatedBundles, newGroup)
	}

	err := AggregatedBundles(aggregatedBundles).Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return aggregatedBundles, nil
}

func (a *Aggregator) aggregatedBundlesMissVersionBundle(bundles, newGroup []Bundle) bool {
	// delta is the number of version bundles diverging compared to the aggregated
	// list of version bundles making up a release. In case delta is greater than
	// 0, the release misses delta version bundles. In case delta is lower than 0,
	// the release has delta version bundles too much. The latter should be way
	// more improbable to happen.
	var delta int
	{
		desireCount := distinctCount(bundles)
		currentCount := len(newGroup)

		delta = desireCount - currentCount
	}

	if delta != 0 {
		v, err := aggregateReleaseVersion(newGroup)
		if err != nil {
			a.logger.Log("level", "error", "message", "failed aggregating release version", "stack", fmt.Sprintf("%#v", err))
		} else {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("release misses %d version bundles", delta), "version", v)
		}

		return true
	}

	return false
}

func (a *Aggregator) bundlesConflictWithDependencies(b1, b2 Bundle) bool {
	for _, d := range b1.Dependencies {
		for _, c := range b2.Components {
			if d.Name != c.Name {
				continue
			}

			if !d.Matches(c) {
				a.logger.Log("component", fmt.Sprintf("%#v", c), "dependency", fmt.Sprintf("%#v", d), "level", "debug", "message", "dependency conflicts with component")
				return true
			}
		}
	}

	return false
}

func (a *Aggregator) containsAggregatedBundle(list [][]Bundle, newGroup []Bundle) bool {
	if len(newGroup) == 0 {
		a.logger.Log("level", "warning", "message", "release aggregation observed empty list of version bundles")
		return false
	}

	for _, grouped := range list {
		if reflect.DeepEqual(grouped, newGroup) {
			v, err := aggregateReleaseVersion(newGroup)
			if err != nil {
				a.logger.Log("level", "error", "message", "failed aggregating release version", "stack", fmt.Sprintf("%#v", err))
			} else {
				a.logger.Log("level", "debug", "message", "release already exists in release list", "version", v)
			}

			return true
		}
	}

	return false
}

func (a *Aggregator) containsBundleByName(list []Bundle, item Bundle) bool {
	for _, b := range list {
		if b.Name == item.Name {
			a.logger.Log("level", "debug", "message", "version bundle already exists in aggregated list", "name", item.Name)
			return true
		}
	}

	return false
}

func distinctCount(list []Bundle) int {
	m := map[string]struct{}{}

	for _, b := range list {
		m[b.Name] = struct{}{}
	}

	return len(m)
}
