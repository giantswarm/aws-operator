package cloudtags

import "context"

type Interface interface {
	// AreClusterTagsEquals compares current cluster tags with the input
	AreClusterTagsEquals(ctx context.Context, clusterID string, tags map[string]string) (bool, error)
	// Get Labels from cluster API object
	GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
}
