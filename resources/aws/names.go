package aws

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func BucketName(cluster awstpr.CustomObject) string {
	return cluster.Spec.Cluster.Customer.ID
}

func BucketObjectDirPath(cluster awstpr.CustomObject) string {
	clusterID := cluster.Spec.Cluster.Cluster.ID
	return fmt.Sprintf("%s/cloudconfig", clusterID)
}

func BucketObjectFullDirPath(cluster awstpr.CustomObject) string {
	bucketName := BucketName(cluster)
	dirPath := BucketObjectDirPath(cluster)
	return fmt.Sprintf("%s/%s", bucketName, dirPath)
}

func BucketObjectName(cluster awstpr.CustomObject, prefix string) string {
	dirPath := BucketObjectDirPath(cluster)
	return fmt.Sprintf("%s/%s", dirPath, prefix)
}
