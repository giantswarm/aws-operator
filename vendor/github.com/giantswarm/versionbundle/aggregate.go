package versionbundle

import (
	"fmt"
	"sort"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

/*
Core design behind Aggregate() implementation:

Aggregate() function takes list of bundles and it builds all possible
combinations of them. Only restrictions are possibly conflicting Bundle
dependencies.

Aggregate() implementation works in three phases:

Assume input [simplified] bundles:
  []Bundle{ (A, 1), (A, 2), (B, 1), (B, 2), (B, 3), (C, 1), (C, 2)}

1. Group all bundles by name and remove duplicate versions.

This produces following map:
  {
	  "A": [(A, 1), (A, 2)]
	  "B": [(B, 1), (B, 2), (B, 3)]
	  "C": [(C, 1), (C, 2)]
  }

2. Build a layered tree from Bundles

                                        (root)
                                          |
                    +---------------------+---------------------+
                    |                                           |
                    |                                           |
                  (A,1)                                       (A,2)
                    |                                           |
      +-------------+-------------+               +-------------+-------------+
      |             |             |               |             |             |
    (B,1)         (B,2)         (B,3)           (B,1)         (B,2)         (B,3)
      |             |             |               |             |             |
  +---+---+     +---+---+     +---+---+       +---+---+     +---+---+     +---+---+
  |       |     |       |     |       |       |       |     |       |     |       |
(C,1)   (C,2) (C,1)   (C,2) (C,1)   (C,2)   (C,1)   (C,2) (C,1)   (C,2) (C,1)   (C,2)


3. Walk the tree and create aggregated bundles where are no dependency
   conflicts.

*/

type AggregatorConfig struct {
	Logger micrologger.Logger
}

type Aggregator struct {
	logger micrologger.Logger
}

// node is a data structure for bundle aggregation tree
type node struct {
	bundle Bundle
	leaves []*node
}

// NewAggregator constructs a new aggregator for given AggregatorConfig.
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

	bundleMap := make(map[string][]Bundle)

	// First group all bundles by name.
	for _, v := range bundles {
		bundleMap[v.Name] = append(bundleMap[v.Name], v)
	}

	// Gather keys for sorting to guarantee always the same order for [][]Bundles
	var keys []string

	// Ensure that there are no duplicate bundles.
	for k, v := range bundleMap {
		sort.Sort(SortBundlesByVersion(v))
		for i := 0; i < len(v)-1; i++ {
			if v[i].Version == v[i+1].Version {
				v = append(v[:i], v[i+1:]...)
				i--
			}
		}
		bundleMap[k] = v
		keys = append(keys, k)
	}

	// Sort'em
	sort.Strings(keys)

	// Tree root is an empty node.
	tree := &node{}

	// Build the tree.
	for _, k := range keys {
		a.walkTreeAndAddLeaves(tree, bundleMap[k])
	}

	// Walk the tree and aggregate.
	aggregatedBundles = a.walkTreeAndAggregate(tree, []Bundle{})

	// Instead of returning empty slice, return explicit nil to be backwards
	// compatible with API.
	if len(aggregatedBundles) == 0 {
		aggregatedBundles = nil
	}

	err := AggregatedBundles(aggregatedBundles).Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return aggregatedBundles, nil
}

func (a *Aggregator) walkTreeAndAddLeaves(n *node, bundles []Bundle) {
	if len(n.leaves) == 0 {
		for _, b := range bundles {
			n.leaves = append(n.leaves, &node{bundle: b})
		}
		return
	}

	for _, leaf := range n.leaves {
		a.walkTreeAndAddLeaves(leaf, bundles)
	}
}

func (a *Aggregator) walkTreeAndAggregate(n *node, bundles []Bundle) [][]Bundle {
	// If current node is leaf, then return bundle if there are no conflicts.
	if len(n.leaves) == 0 {
		// Only aggregate bundle groups that don't have conflicting dependencies.
		for i, b1 := range bundles {
			for j, b2 := range bundles {
				// No need to self-verify.
				if i == j {
					continue
				}

				if a.bundlesConflictWithDependencies(b1, b2) {
					return [][]Bundle{}
				}
			}
		}

		sort.Sort(SortBundlesByVersion(bundles))
		sort.Stable(SortBundlesByName(bundles))
		return [][]Bundle{bundles}
	}

	// In the middle of the tree -> continue walking.
	aggregates := make([][]Bundle, 0)
	for _, leaf := range n.leaves {
		bundlesCopy := make([]Bundle, len(bundles), len(bundles)+1)
		copy(bundlesCopy, bundles)
		bundlesCopy = append(bundlesCopy, leaf.bundle)
		aggregates = append(aggregates, a.walkTreeAndAggregate(leaf, bundlesCopy)...)
	}

	return aggregates
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

	for _, d := range b2.Dependencies {
		for _, c := range b1.Components {
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
