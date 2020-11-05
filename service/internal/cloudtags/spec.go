package cloudtags

import (
	"context"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Interface interface {
	// CloudTagsNotInSync compares current cluster labels with the CF stack tags
	CloudTagsNotInSync(ctx context.Context, getter key.LabelsGetter, stackType string) (bool, error)
	// Get Labels from cluster API object
	GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error)
}
