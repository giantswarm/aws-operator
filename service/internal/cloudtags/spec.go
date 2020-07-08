package cloudtags

import "context"

type Interface interface {
	// ClusterLabelsNotEqual compares current cluster labels with the stack tags passed
	ClusterLabelsNotEqual(ctx context.Context, clusterID string, stags map[string]string) (bool, error)
	// Get Labels from cluster API object
	GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
}
