package alerter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindOrphanResources(t *testing.T) {
	tests := []struct {
		clusterIDs      []string
		resourceNames   []string
		orphanResources []string
	}{
		// Case 1. No resources.
		{
			clusterIDs:      []string{},
			resourceNames:   []string{},
			orphanResources: []string{},
		},
		// Case 2. Cluster and resources match.
		{
			clusterIDs: []string{
				"cluster-1",
				"cluster-2",
			},
			resourceNames: []string{
				"cluster-1",
				"cluster-2",
			},
			orphanResources: []string{},
		},
		// Case 3. Orphan resources.
		{
			clusterIDs: []string{
				"cluster-1",
			},
			resourceNames: []string{
				"cluster-1",
				"cluster-2",
				"cluster-3",
			},
			orphanResources: []string{
				"cluster-2",
				"cluster-3",
			},
		},
		// Case 3. Single orphan resource.
		{
			clusterIDs: []string{
				"cluster-1",
			},
			resourceNames: []string{
				"cluster-1",
				"cluster-2",
			},
			orphanResources: []string{
				"cluster-2",
			},
		},
		// Case 4. Don't alert on missing resource as it may not have been
		//created yet.
		{
			clusterIDs: []string{
				"cluster-1",
			},
			resourceNames:   []string{},
			orphanResources: []string{},
		},
	}

	for _, tc := range tests {
		results := FindOrphanResources(tc.clusterIDs, tc.resourceNames)
		assert.EqualValues(t, tc.orphanResources, results)
	}
}
