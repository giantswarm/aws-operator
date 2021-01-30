package cloudtags

import (
	"context"
)

type Interface interface {
	// Get Labels from cluster API object
	GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
}
