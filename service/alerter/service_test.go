package alerter

import (
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Test_FindOrphanClusters(t *testing.T) {
	n := time.Now()
	twentyMinAgo := metav1.Time{n.Add(-20 * time.Minute)}
	fiveMinAgo := metav1.Time{n.Add(-5 * time.Minute)}
	tests := []struct {
		clusters       []v1alpha1.AWSConfig
		resourceNames  []string
		orphanClusters []string
		description    string
	}{
		{
			description:    "all empty, expected empty",
			clusters:       []v1alpha1.AWSConfig{},
			resourceNames:  []string{},
			orphanClusters: []string{},
		},
		{
			description:    "orphan resources, empty clusters, expected empty orphan clusters",
			clusters:       []v1alpha1.AWSConfig{},
			resourceNames:  []string{"id1", "id2"},
			orphanClusters: []string{},
		},
		{
			description: "matching clusters and resources, expected empty",
			clusters: []v1alpha1.AWSConfig{
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id1",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id2",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
			},
			resourceNames:  []string{"id1", "id2"},
			orphanClusters: []string{},
		},
		{
			description: "orphan cluster old enough, expected to be reported",
			clusters: []v1alpha1.AWSConfig{
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id1",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id2",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id3",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
			},
			resourceNames:  []string{"id1", "id2"},
			orphanClusters: []string{"id3"},
		},
		{
			description: "recent orphan cluster, not expected to be reported",
			clusters: []v1alpha1.AWSConfig{
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id1",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: fiveMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id2",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: fiveMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id3",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: fiveMinAgo,
					},
				},
			},
			resourceNames:  []string{"id1", "id2"},
			orphanClusters: []string{},
		},
		{
			description: "multiple orphan resources and cluster, both recent and old",
			clusters: []v1alpha1.AWSConfig{
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id1",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id2",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: fiveMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id3",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id4",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id5",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: fiveMinAgo,
					},
				},
				{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "id6",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: twentyMinAgo,
					},
				},
			},
			resourceNames:  []string{"id1", "id2", "id5", "id7"},
			orphanClusters: []string{"id3", "id4", "id6"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			results := FindOrphanClusters(tc.clusters, tc.resourceNames)
			assert.EqualValues(t, tc.orphanClusters, results)
		})
	}
}
