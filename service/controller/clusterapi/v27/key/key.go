package key

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use. It is used
	// in the main stack output and S3 object paths.
	CloudConfigVersion = "v_4_0_0"
)

const (
	EC2RoleK8s = "EC2-K8S-Role"
)

const (
	LabelApp           = "app"
	LabelCluster       = "giantswarm.io/cluster"
	LabelOrganization  = "giantswarm.io/organization"
	LabelVersionBundle = "giantswarm.io/version-bundle"
)

// AWS Tags used for cost analysis and general resource tagging.
const (
	TagCluster      = "giantswarm.io/cluster"
	TagInstallation = "giantswarm.io/installation"
	TagOrganization = "giantswarm.io/organization"
)

func AutoScalingGroupName(cluster v1alpha1.Cluster, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(cluster), groupName)
}

func BucketName(cluster v1alpha1.Cluster, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", accountID, ClusterID(cluster))
}

// BucketObjectName computes the S3 object path to the actual cloud config.
//
//     /version/3.4.0/cloudconfig/v_3_2_5/master
//     /version/3.4.0/cloudconfig/v_3_2_5/worker
//
func BucketObjectName(cluster v1alpha1.Cluster, role string) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s", VersionBundleVersion(cluster), CloudConfigVersion, role)
}

func ClusterAPIEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("api.%s.%s", ClusterID(cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Cluster.DNS.Domain
}

func ClusterCloudProviderTag(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("kubernetes.io/cluster/%s", ClusterID(cluster))
}

func ClusterEtcdEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("etcd.%s.%s:2379", ClusterID(cluster), ClusterBaseDomain(cluster))
}

func ClusterID(cluster v1alpha1.Cluster) string {
	return providerStatus(cluster).Cluster.ID
}

func ClusterNamespace(cluster v1alpha1.Cluster) string {
	return ClusterID(cluster)
}

func ClusterTags(cluster v1alpha1.Cluster, installationName string) map[string]string {
	TagCloudProvider := ClusterCloudProviderTag(cluster)

	tags := map[string]string{
		TagCloudProvider: "owned",
		TagCluster:       ClusterID(cluster),
		TagInstallation:  installationName,
		TagOrganization:  OrganizationID(cluster),
	}

	return tags
}

func CredentialName(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Provider.CredentialSecret.Name
}

func CredentialNamespace(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Provider.CredentialSecret.Namespace
}

func DockerVolumeResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("DockerVolume", cluster)
}

func EC2ServiceDomain(cluster v1alpha1.Cluster) string {
	domain := "ec2.amazonaws.com"

	if isChinaRegion(cluster) {
		domain += ".cn"
	}

	return domain
}

func ELBNameAPI(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-api", ClusterID(cluster))
}

func ELBNameEtcd(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-etcd", ClusterID(cluster))
}

func ELBNameIngress(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-ingress", ClusterID(cluster))
}

func MasterInstanceResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("MasterInstance", cluster)
}

func OrganizationID(cluster v1alpha1.Cluster) string {
	return cluster.Labels[LabelOrganization]
}

func ProfileName(cluster v1alpha1.Cluster, profileType string) string {
	return RoleName(cluster, profileType)
}

func Region(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Provider.Region
}

func RegionARN(cluster v1alpha1.Cluster) string {
	regionARN := "aws"

	if isChinaRegion(cluster) {
		regionARN += "-cn"
	}

	return regionARN
}

func RoleName(cluster v1alpha1.Cluster, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(cluster), profileType, EC2RoleK8s)
}

func SmallCloudConfigPath(cluster v1alpha1.Cluster, accountID string, role string) string {
	return fmt.Sprintf("%s/%s", BucketName(cluster, accountID), BucketObjectName(cluster, role))
}

func SmallCloudConfigS3URL(cluster v1alpha1.Cluster, accountID string, role string) string {
	return fmt.Sprintf("s3://%s", SmallCloudConfigPath(cluster, accountID, role))
}

func StackNameCPF(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("cluster-%s-host-main", ClusterID(cluster))
}

func StackNameCPI(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("cluster-%s-host-setup", ClusterID(cluster))
}

func StackNameTCCP(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("cluster-%s-guest-main", ClusterID(cluster))
}

// VersionBundleVersion returns the version contained in the Version Bundle.
func VersionBundleVersion(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Cluster.VersionBundle.Version
}

func VolumeNameDocker(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-docker", ClusterID(cluster))
}

func VolumeNameEtcd(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-etcd", ClusterID(cluster))
}

func VolumeNameLog(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-log", ClusterID(cluster))
}

func baseRoleARN(cluster v1alpha1.Cluster, accountID string, kind string) string {
	clusterID := ClusterID(cluster)
	partition := RegionARN(cluster)

	return fmt.Sprintf("arn:%s:iam::%s:role/%s-%s-%s", partition, accountID, clusterID, kind, EC2RoleK8s)
}

// getResourcenameWithTimeHash returns a string cromprised of some prefix, a
// time hash and a cluster ID.
func getResourcenameWithTimeHash(prefix string, cluster v1alpha1.Cluster) string {
	id := strings.Replace(ClusterID(cluster), "-", "", -1)

	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(id)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}

func isChinaRegion(cluster v1alpha1.Cluster) bool {
	return strings.HasPrefix(Region(cluster), "cn-")
}
