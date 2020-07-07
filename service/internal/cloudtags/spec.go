package cloudtags

import "context"

type Interface interface {
	// AreClusterTagsEquals compares current cluster tags with the stack tags
	AreClusterTagsEquals(ctx context.Context, ctags map[string]string, stags map[string]string) bool
	// Get Labels from cluster API object
	GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
}
