package keyv2

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	// CloudProviderTagName is used to add Cloud Provider tags to AWS resources.
	CloudProviderTagName = "kubernetes.io/cluster/%s"

	// CloudProviderTagOwnedValue is used to indicate an AWS resource is owned
	// and managed by a cluster.
	CloudProviderTagOwnedValue = "owned"

	// LegacyVersion is the version in the version bundle for existing clusters.
	LegacyVersion = "0.1.0"

	// ProfileNameTemplate will be included in the IAM instance profile name.
	ProfileNameTemplate = "EC2-K8S-Role"
	// RoleNameTemplate will be included in the IAM role name.
	RoleNameTemplate = "EC2-K8S-Role"
	// PolicyNameTemplate will be included in the IAM policy name.
	PolicyNameTemplate = "EC2-K8S-Policy"
)

func AutoScalingGroupName(customObject v1alpha1.AWSConfig, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func AvailabilityZone(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.AZ
}

func BucketName(customObject v1alpha1.AWSConfig, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", accountID, ClusterID(customObject))
}

func BucketObjectName(templateVersion string, prefix string) string {
	return fmt.Sprintf("cloudconfig/%s/%s", templateVersion, prefix)
}

func ClusterCustomer(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.ID
}

func ClusterNamespace(customObject v1alpha1.AWSConfig) string {
	return ClusterID(customObject)
}

func ClusterTags(customObject v1alpha1.AWSConfig) map[string]string {
	cloudProviderTag := fmt.Sprintf(CloudProviderTagName, ClusterID(customObject))
	tags := map[string]string{
		cloudProviderTag: CloudProviderTagOwnedValue,
	}

	return tags
}

func CustomerID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterVersion(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Version
}

func HasClusterVersion(customObject v1alpha1.AWSConfig) bool {
	switch ClusterVersion(customObject) {
	case string("v_0_1_0"):
		return true
	default:
		return false
	}
}

func IngressControllerInsecurePort(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Kubernetes.IngressController.InsecurePort
}

func IngressControllerSecurePort(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Kubernetes.IngressController.SecurePort
}

func InstanceProfileName(customObject v1alpha1.AWSConfig, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(customObject), profileType, ProfileNameTemplate)
}

func KubernetesAPISecurePort(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Kubernetes.API.SecurePort
}

// LoadBalancerName produces a unique name for the load balancer.
// It takes the domain name, extracts the first subdomain, and combines it with the cluster name.
func LoadBalancerName(domainName string, cluster v1alpha1.AWSConfig) (string, error) {
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

func MainGuestStackName(customObject v1alpha1.AWSConfig) string {
	clusterID := ClusterID(customObject)

	return fmt.Sprintf("cluster-%s-guest-main", clusterID)
}

func MainHostPreStackName(customObject v1alpha1.AWSConfig) string {
	clusterID := ClusterID(customObject)

	return fmt.Sprintf("cluster-%s-host-setup", clusterID)
}

func MainHostPostStackName(customObject v1alpha1.AWSConfig) string {
	clusterID := ClusterID(customObject)

	return fmt.Sprintf("cluster-%s-host-main", clusterID)
}

func MasterCount(customObject v1alpha1.AWSConfig) int {
	return len(customObject.Spec.AWS.Masters)
}

func MasterImageID(customObject v1alpha1.AWSConfig) string {
	var imageID string

	if len(customObject.Spec.AWS.Masters) > 0 {
		imageID = customObject.Spec.AWS.Masters[0].ImageID
	}

	return imageID
}

func MasterInstanceName(customObject v1alpha1.AWSConfig) string {
	clusterID := ClusterID(customObject)

	return fmt.Sprintf("%s-master", clusterID)
}

func MasterInstanceType(customObject v1alpha1.AWSConfig) string {
	var instanceType string

	if len(customObject.Spec.AWS.Masters) > 0 {
		instanceType = customObject.Spec.AWS.Masters[0].InstanceType
	}

	return instanceType
}

func PeerAccessRoleName(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(customObject))
}

func PeerID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.VPC.PeerID
}

func PolicyName(customObject v1alpha1.AWSConfig, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(customObject), profileType, PolicyNameTemplate)
}

func PrivateSubnetCIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.VPC.PrivateSubnetCIDR
}

func Region(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.Region
}

func RoleName(customObject v1alpha1.AWSConfig, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(customObject), profileType, RoleNameTemplate)
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

func RouteTableName(customObject v1alpha1.AWSConfig, suffix string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), suffix)
}

func SecurityGroupName(customObject v1alpha1.AWSConfig, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func SubnetName(customObject v1alpha1.AWSConfig, suffix string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), suffix)
}

func ToCustomObject(v interface{}) (v1alpha1.AWSConfig, error) {
	if v == nil {
		return v1alpha1.AWSConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.AWSConfig{}, v)
	}

	customObjectPointer, ok := v.(*v1alpha1.AWSConfig)
	if !ok {
		return v1alpha1.AWSConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.AWSConfig{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

// VersionBundleVersion returns the version contained in the Version Bundle.
func VersionBundleVersion(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.VersionBundle.Version
}

func WorkerCount(customObject v1alpha1.AWSConfig) int {
	return len(customObject.Spec.AWS.Workers)
}

func WorkerImageID(customObject v1alpha1.AWSConfig) string {
	var imageID string

	if len(customObject.Spec.AWS.Workers) > 0 {
		imageID = customObject.Spec.AWS.Workers[0].ImageID
	}

	return imageID
}

func WorkerInstanceType(customObject v1alpha1.AWSConfig) string {
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

// ImageID returns the EC2 AMI for the configured region.
func ImageID(customObject v1alpha1.AWSConfig) (string, error) {
	region := Region(customObject)

	/*
		Container Linux AMIs for each active AWS region.
		From: https://coreos.com/os/docs/latest/booting-on-ec2.html

		NOTE 1: AMIs should always be for HVM virtualisation and not PV.
		NOTE 2: You also need to update the tests.

		service/keyv2/key_test.go
		service/resource/cloudformationv2/adapter/adapter_test.go
		service/resource/cloudformationv2/main_stack_test.go

		Current Release: CoreOS Container Linux stable 1576.5.0 (HVM)
	*/
	imageIDs := map[string]string{
		"eu-central-1": "ami-90c152ff",
		"eu-west-1":    "ami-32d1474b",
		"us-west-2":    "ami-dc4ce6a4",
	}

	imageID, ok := imageIDs[region]
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "no image id for region '%s'", region)
	}

	return imageID, nil
}
