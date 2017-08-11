package key

import (
	"github.com/giantswarm/awstpr"
)

func AvailabilityZone(customObject awstpr.CustomObject) string {
	return customObject.Spec.AWS.AZ
}

func ClusterID(customObject awstpr.CustomObject) string {
	return customObject.Spec.Cluster.Cluster.ID
}

func WorkerCount(customObject awstpr.CustomObject) int {
	return len(customObject.Spec.AWS.Workers)
}
