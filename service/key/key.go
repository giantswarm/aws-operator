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

func MasterImageID(customObject awstpr.CustomObject) string {
	var imageID string

	if len(customObject.Spec.AWS.Masters) > 0 {
		imageID = customObject.Spec.AWS.Masters[0].ImageID
	}

	return imageID
}

func MasterInstanceType(customObject awstpr.CustomObject) string {
	var instanceType string

	if len(customObject.Spec.AWS.Masters) > 0 {
		instanceType = customObject.Spec.AWS.Masters[0].InstanceType
	}

	return instanceType
}

func SecurityGroupName(customObject awstpr.CustomObject, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func WorkerCount(customObject awstpr.CustomObject) int {
	return len(customObject.Spec.AWS.Workers)
}

func WorkerImageID(customObject awstpr.CustomObject) string {
	var imageID string

	if len(customObject.Spec.AWS.Workers) > 0 {
		imageID = customObject.Spec.AWS.Workers[0].ImageID
	}

	return imageID
}

func WorkerInstanceType(customObject awstpr.CustomObject) string {
	var instanceType string

	if len(customObject.Spec.AWS.Workers) > 0 {
		instanceType = customObject.Spec.AWS.Workers[0].InstanceType

	}

	return instanceType
}
