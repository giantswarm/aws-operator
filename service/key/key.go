package key

import (
	"fmt"

	"github.com/giantswarm/awstpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	"github.com/giantswarm/microerror"
)

const (
	// cloudFormationVersion is set during the migration so resources are
	// managed by the cloudformation resource.
	// TODO Remove once the migration is complete.
	cloudFormationVersion = "cloud-formation"

	// ProfileNameTemplate will be included in the IAM instance profile name.
	ProfileNameTemplate = "EC2-K8S-Role"
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

func CustomerID(customObject awstpr.CustomObject) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterVersion(customObject awstpr.CustomObject) string {
	return customObject.Spec.Cluster.Version
}

func HasClusterVersion(customObject awstpr.CustomObject) bool {
	switch ClusterVersion(customObject) {
	case string(cloudconfig.V_0_1_0):
		return true
	default:
		return false
	}
}

func InstanceProfileName(customObject awstpr.CustomObject, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(customObject), profileType, ProfileNameTemplate)
}

func MainStackName(customObject awstpr.CustomObject) string {
	clusterID := ClusterID(customObject)

	return fmt.Sprintf("%s-main", clusterID)
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

func RouteTableName(customObject awstpr.CustomObject, suffix string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), suffix)
}

func SecurityGroupName(customObject awstpr.CustomObject, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func SubnetName(customObject awstpr.CustomObject, suffix string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), suffix)
}

func ToCustomObject(v interface{}) (awstpr.CustomObject, error) {
	if v == nil {
		return awstpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &awstpr.CustomObject{}, v)
	}

	customObjectPointer, ok := v.(*awstpr.CustomObject)
	if !ok {
		return awstpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &awstpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

// UseCloudFormation returns true if the cluster version is "cloud-formation".
// This will be used while we migrate to Cloud Formation and then removed.
func UseCloudFormation(customObject awstpr.CustomObject) bool {
	if ClusterVersion(customObject) == cloudFormationVersion {
		return true
	}

	return false
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
