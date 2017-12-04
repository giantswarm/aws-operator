package key

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/giantswarm/awstpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	"github.com/giantswarm/microerror"
)

const (
	// CloudFormationVersion is the version in the version bundle for
	// transitioning to Cloud Formation.
	// TODO Remove once the migration is complete.
	CloudFormationVersion = "0.2.0"

	// LegacyVersion is the version in the version bundle for existing clusters.
	LegacyVersion = "0.1.0"

	// ProfileNameTemplate will be included in the IAM instance profile name.
	ProfileNameTemplate = "EC2-K8S-Role"
)

func AutoScalingGroupName(customObject awstpr.CustomObject, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func AvailabilityZone(customObject awstpr.CustomObject) string {
	return customObject.Spec.AWS.AZ
}

func BucketName(customObject awstpr.CustomObject, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", ClusterID(customObject), accountID)
}

func ClusterCustomer(customObject awstpr.CustomObject) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterID(customObject awstpr.CustomObject) string {
	return customObject.Spec.Cluster.Cluster.ID
}

func ClusterNamespace(customObject awstpr.CustomObject) string {
	return ClusterID(customObject)
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

// LoadBalancerName produces a unique name for the load balancer.
// It takes the domain name, extracts the first subdomain, and combines it with the cluster name.
func LoadBalancerName(domainName string, cluster awstpr.CustomObject) (string, error) {
	if ClusterID(cluster) == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	componentName, err := componentName(domainName)
	if err != nil {
		return "", microerror.Maskf(malformedCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	lbName := fmt.Sprintf("%s-%s", ClusterID(cluster), componentName)

	return lbName, nil
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

// UseCloudFormation returns true if the version in the version bundle matches
// the Cloud Formation version.
// TODO Remove once we've migrated all AWS resources to Cloud Formation.
func UseCloudFormation(customObject awstpr.CustomObject) bool {
	if VersionBundleVersion(customObject) == CloudFormationVersion {
		return true
	}

	return false
}

// VersionBundleVersion returns the version contained in the Version Bundle.
func VersionBundleVersion(customObject awstpr.CustomObject) string {
	return customObject.Spec.VersionBundle.Version
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

// componentName returns the first component of a domain name.
// e.g. apiserver.example.customer.cloud.com -> apiserver
func componentName(domainName string) (string, error) {
	splits := strings.SplitN(domainName, ".", 2)

	if len(splits) != 2 {
		return "", microerror.Mask(malformedCloudConfigKeyError)
	}

	return splits[0], nil
}

// RootDir returns the path in the base directory until the
// root elemant is found.
func RootDir(baseDir, rootElement string) (string, error) {
	items := strings.Split(baseDir, string(filepath.Separator))
	rootIndex := -1
	for i := len(items) - 1; i >= 0; i-- {
		if items[i] == rootElement {
			rootIndex = i
			break
		}
	}
	if rootIndex == -1 {
		return "", microerror.Mask(notFoundError)
	}

	return "/" + filepath.Join(items[:(rootIndex+1)]...), nil
}
