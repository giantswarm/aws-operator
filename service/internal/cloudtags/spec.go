package cloudtags

import "context"

type Interface interface {
	// CloudTagsNotInSync compares current cluster labels with the CF stack tags
	CloudTagsNotInSync(ctx context.Context, cr interface{}, stackType string) (bool, error)
	// Get Labels from cluster API object
	GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
	// Get Labels from AWS Cloud Formation Stack
	GetAWSTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
}
