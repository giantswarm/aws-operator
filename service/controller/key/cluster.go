package key

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use. It is used
	// in the main stack output and S3 object paths.
	CloudConfigVersion = "v_4_8_0"
	CloudProvider      = "aws"
)

const (
	EC2RoleK8s   = "EC2-K8S-Role"
	EC2PolicyK8s = "EC2-K8S-Policy"
)

const (
	EtcdPort             = 2379
	EtcdPrefix           = "giantswarm.io"
	KubernetesSecurePort = 443
)

// AWS Tags used for cost analysis and general resource tagging.
const (
	TagAvailabilityZone  = "giantswarm.io/availability-zone"
	TagCluster           = "giantswarm.io/cluster"
	TagInstallation      = "giantswarm.io/installation"
	TagMachineDeployment = "giantswarm.io/machine-deployment"
	TagOrganization      = "giantswarm.io/organization"
	TagRouteTableType    = "giantswarm.io/route-table-type"
	TagStack             = "giantswarm.io/stack"
	TagSubnetType        = "giantswarm.io/subnet-type"
)

const (
	StackTCCP  = "tccp"
	StackTCCPF = "tccpf"
	StackTCCPI = "tccpi"
	StackTCNP  = "tcnp"
	StackTCNPF = "tcnpf"
)

const (
	RefNodeDrainer = "NodeDrainer"
	RefWorkerASG   = "workerAutoScalingGroup"
)

func ClusterAPIEndpoint(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("api.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.DNS.Domain
}

func ClusterEtcdEndpoint(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("etcd.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterEtcdEndpointWithPort(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s:2379", ClusterEtcdEndpoint(cluster))
}

func ClusterKubeletEndpoint(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("worker.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterNamespace(cluster infrastructurev1alpha2.Cluster) string {
	return ClusterID(&cluster)
}

func CredentialName(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Name
}

func CredentialNamespace(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Namespace
}

func DockerVolumeResourceName(cr infrastructurev1alpha2.Cluster, t time.Time) string {
	return getResourcenameWithTimeHash("DockerVolume", cr, t)
}

func MasterAvailabilityZone(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.AvailabilityZone
}

func MasterCount(cluster infrastructurev1alpha2.Cluster) int {
	return 1
}

func MasterInstanceResourceName(cr infrastructurev1alpha2.Cluster, t time.Time) string {
	return getResourcenameWithTimeHash("MasterInstance", cr, t)
}

func MasterInstanceName(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-master", ClusterID(&cluster))
}

func MasterInstanceType(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.InstanceType
}

func ManagedRecordSets(cluster infrastructurev1alpha2.Cluster) []string {
	tcBaseDomain := TenantClusterBaseDomain(cluster)
	return []string{
		fmt.Sprintf("%s.", tcBaseDomain),
		fmt.Sprintf("\\052.%s.", tcBaseDomain), // \\052 - `*` wildcard record
		fmt.Sprintf("api.%s.", tcBaseDomain),
		fmt.Sprintf("etcd.%s.", tcBaseDomain),
		fmt.Sprintf("internal-api.%s.", tcBaseDomain),
	}
}

func OIDCClientID(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.OIDC.ClientID
}
func OIDCIssuerURL(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.OIDC.IssuerURL
}
func OIDCUsernameClaim(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.OIDC.Claims.Username
}
func OIDCGroupsClaim(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.OIDC.Claims.Groups
}

func PolicyNameMaster(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(&cluster), EC2PolicyK8s)
}

func ProfileNameMaster(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(&cluster), EC2RoleK8s)
}

func Region(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Region
}

func RoleNameMaster(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(&cluster), EC2RoleK8s)
}

func RolePeerAccess(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(&cluster))
}

func RouteTableName(cluster infrastructurev1alpha2.Cluster, suffix, az string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(&cluster), suffix, az)
}

func StatusClusterNetworkCIDR(cluster infrastructurev1alpha2.Cluster) string {
	return clusterProviderStatus(cluster).Provider.Network.CIDR
}

func TargetLogBucketName(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-g8s-access-logs", ClusterID(&cluster))
}

func TenantClusterBaseDomain(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ToCluster(v interface{}) (infrastructurev1alpha2.Cluster, error) {
	if v == nil {
		return infrastructurev1alpha2.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.Cluster{}, v)
	}

	p, ok := v.(*infrastructurev1alpha2.Cluster)
	if !ok {
		return infrastructurev1alpha2.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.Cluster{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}

func VolumeNameDocker(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-docker", ClusterID(&cluster))
}

func VolumeNameEtcd(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-etcd", ClusterID(&cluster))
}

func VolumeNameLog(cluster infrastructurev1alpha2.Cluster) string {
	return fmt.Sprintf("%s-log", ClusterID(&cluster))
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
func getResourcenameWithTimeHash(prefix string, cluster infrastructurev1alpha2.Cluster, t time.Time) string {
	id := strings.Replace(ClusterID(&cluster), "-", "", -1)

	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(t.UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(id)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}
