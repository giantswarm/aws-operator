package key

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use. It is used
	// in the main stack output and S3 object paths.
	CloudConfigVersion = "v_4_0_0"
	CloudProvider      = "aws"
)

const (
	EC2RoleK8s   = "EC2-K8S-Role"
	EC2PolicyK8s = "EC2-K8S-Policy"
)

const (
	IngressControllerInsecurePort = 30010
	IngressControllerSecurePort   = 30011
)

const (
	EtcdPort             = 2379
	EtcdPrefix           = "giantswarm.io"
	KubernetesSecurePort = 443
)

// AWS Tags used for cost analysis and general resource tagging.
const (
	TagCluster      = "giantswarm.io/cluster"
	TagInstallation = "giantswarm.io/installation"
	TagOrganization = "giantswarm.io/organization"
)

const (
	RefNodeDrainer = "NodeDrainer"
	RefWorkerASG   = "workerAutoScalingGroup"
)

func BucketName(cluster v1alpha1.Cluster, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", accountID, ClusterID(cluster))
}

// BucketObjectName computes the S3 object path to the actual cloud config.
//
//     /version/3.4.0/cloudconfig/v_3_2_5/master
//     /version/3.4.0/cloudconfig/v_3_2_5/worker
//
func BucketObjectName(cluster v1alpha1.Cluster, role string) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s", OperatorVersion(&cluster), CloudConfigVersion, role)
}

func ClusterAPIEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("api.%s.%s", ClusterID(cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.DNS.Domain
}

func ClusterCloudProviderTag(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("kubernetes.io/cluster/%s", ClusterID(cluster))
}

func ClusterEtcdEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("etcd.%s.%s", ClusterID(cluster), ClusterBaseDomain(cluster))
}

func ClusterEtcdEndpointWithPort(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s:2379", ClusterEtcdEndpoint(cluster))
}

func ClusterID(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Cluster.ID
}

func ClusterKubeletEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("worker.%s.%s", ClusterID(cluster), ClusterBaseDomain(cluster))
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
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Name
}

func CredentialNamespace(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Namespace
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

func ImageID(cluster v1alpha1.Cluster) string {
	return imageIDs()[Region(cluster)]
}

func KubeletLabels(cluster v1alpha1.Cluster) string {
	var labels string

	labels = ensureLabel(labels, label.Provider, "aws")
	labels = ensureLabel(labels, label.ReleaseVersion, ReleaseVersion(&cluster))

	return labels
}

func MasterCount(cluster v1alpha1.Cluster) int {
	return 1
}

func MasterInstanceResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("MasterInstance", cluster)
}

func MasterInstanceName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master", ClusterID(cluster))
}

func MasterInstanceType(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.InstanceType
}

func OrganizationID(cluster v1alpha1.Cluster) string {
	return cluster.Labels[label.Organization]
}

func PolicyNameMaster(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(cluster), EC2PolicyK8s)
}

func PolicyNameWorker(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-worker-%s", ClusterID(cluster), EC2PolicyK8s)
}

func ProfileNameMaster(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(cluster), EC2RoleK8s)
}

func ProfileNameWorker(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-worker-%s", ClusterID(cluster), EC2RoleK8s)
}

func Region(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Region
}

func RegionARN(cluster v1alpha1.Cluster) string {
	regionARN := "aws"

	if isChinaRegion(cluster) {
		regionARN += "-cn"
	}

	return regionARN
}

func RoleARNMaster(cluster v1alpha1.Cluster, accountID string) string {
	return baseRoleARN(cluster, accountID, "master")
}

func RoleARNWorker(cluster v1alpha1.Cluster, accountID string) string {
	return baseRoleARN(cluster, accountID, "worker")
}

func RoleNameMaster(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(cluster), EC2RoleK8s)
}

func RoleNameWorker(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-worker-%s", ClusterID(cluster), EC2RoleK8s)
}

func RolePeerAccess(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(cluster))
}

func RouteTableName(cluster v1alpha1.Cluster, suffix string, idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return fmt.Sprintf("%s-%s", ClusterID(cluster), suffix)
	}
	return fmt.Sprintf("%s-%s%02d", ClusterID(cluster), suffix, idx)
}

func SecurityGroupName(cluster v1alpha1.Cluster, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(cluster), groupName)
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

func StatusClusterNetworkCIDR(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Provider.Network.CIDR
}

func TargetLogBucketName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-g8s-access-logs", ClusterID(cluster))
}

func ToCluster(v interface{}) (v1alpha1.Cluster, error) {
	if v == nil {
		return v1alpha1.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Cluster{}, v)
	}

	p, ok := v.(*v1alpha1.Cluster)
	if !ok {
		return v1alpha1.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Cluster{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
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

func ensureLabel(labels string, key string, value string) string {
	if key == "" {
		return labels
	}
	if value == "" {
		return labels
	}

	var split []string
	if labels != "" {
		split = strings.Split(labels, ",")
	}

	var found bool
	for i, l := range split {
		if !strings.HasPrefix(l, key+"=") {
			continue
		}

		found = true
		split[i] = key + "=" + value
	}

	if !found {
		split = append(split, key+"="+value)
	}

	joined := strings.Join(split, ",")

	return joined
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

// imageIDs returns our Container Linux AMIs for each active AWS region. Note
// that AMIs should always be for HVM virtualisation, not PV. Current Release is
// CoreOS Container Linux stable 2079.5.1. AMI IDs are copied from the following
// resource.
//
//     https://stable.release.core-os.net/amd64-usr/2079.5.1/coreos_production_ami_hvm.txt.
//
func imageIDs() map[string]string {
	return map[string]string{
		"ap-northeast-1": "ami-036857bdeb5b3362c",
		"ap-northeast-2": "ami-00f2ee0ba5c3954e9",
		"ap-south-1":     "ami-060a95c11ed11c1bd",
		"ap-southeast-1": "ami-0da0dfdf36db6e7e1",
		"ap-southeast-2": "ami-0c245ecdf4720b5a2",
		"ca-central-1":   "ami-0867c908d18a8c69e",
		"cn-north-1":     "ami-0032227ab96e75a9f",
		"cn-northwest-1": "ami-006bc343e8c9c9b22",
		"eu-central-1":   "ami-0018c6ee88479b31e",
		"eu-north-1":     "ami-00e03f7974618119a",
		"eu-west-1":      "ami-018351c24af175181",
		"eu-west-2":      "ami-0b91a753d4fa446b4",
		"eu-west-3":      "ami-0719f5491f02f1874",
		"sa-east-1":      "ami-08003539c64a2c6b9",
		"us-east-1":      "ami-0a7247846b022222c",
		"us-east-2":      "ami-0efcfe3a15d87beff",
		"us-gov-east-1":  "ami-0229105a89981165a",
		"us-gov-west-1":  "ami-22384043",
		"us-west-1":      "ami-062f6abca7bac0908",
		"us-west-2":      "ami-0e7d76904282e972b",
	}
}

func isChinaRegion(cluster v1alpha1.Cluster) bool {
	return strings.HasPrefix(Region(cluster), "cn-")
}
