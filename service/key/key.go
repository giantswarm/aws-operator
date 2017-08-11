package key

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func AutoScalingGroupName(customObject awstpr.CustomObject, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func AvailabilityZone(customObject awstpr.CustomObject) string {
	return customObject.Spec.AWS.AZ
}

func ClusterID(customObject awstpr.CustomObject) string {
	return customObject.Spec.Cluster.Cluster.ID
}

func SecurityGroupName(customObject awstpr.CustomObject, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func WorkerCount(customObject awstpr.CustomObject) int {
	return len(customObject.Spec.AWS.Workers)
}
