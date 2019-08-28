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
	CloudConfigVersion = "v_4_7_0"
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

func ClusterAPIEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("api.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.DNS.Domain
}

func ClusterEtcdEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("etcd.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterEtcdEndpointWithPort(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s:2379", ClusterEtcdEndpoint(cluster))
}

func ClusterKubeletEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("worker.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterNamespace(cluster v1alpha1.Cluster) string {
	return ClusterID(&cluster)
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

func KubeletLabels(cluster v1alpha1.Cluster) string {
	var labels string

	labels = ensureLabel(labels, label.Provider, "aws")
	labels = ensureLabel(labels, label.ReleaseVersion, ReleaseVersion(&cluster))

	return labels
}

func MasterAvailabilityZone(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.AvailabilityZone
}

func MasterCount(cluster v1alpha1.Cluster) int {
	return 1
}

func MasterInstanceResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("MasterInstance", cluster)
}

func MasterInstanceName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master", ClusterID(&cluster))
}

func MasterInstanceType(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.InstanceType
}

func PolicyNameMaster(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(&cluster), EC2PolicyK8s)
}

func ProfileNameMaster(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(&cluster), EC2RoleK8s)
}

func Region(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Region
}

func RoleNameMaster(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(&cluster), EC2RoleK8s)
}

func RolePeerAccess(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(&cluster))
}

func RouteTableName(cluster v1alpha1.Cluster, suffix, az string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(&cluster), suffix, az)
}

func StatusClusterNetworkCIDR(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Provider.Network.CIDR
}

func TargetLogBucketName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-g8s-access-logs", ClusterID(&cluster))
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
	return fmt.Sprintf("%s-docker", ClusterID(&cluster))
}

func VolumeNameEtcd(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-etcd", ClusterID(&cluster))
}

func VolumeNameLog(cluster v1alpha1.Cluster) string {
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
func getResourcenameWithTimeHash(prefix string, cluster v1alpha1.Cluster) string {
	id := strings.Replace(ClusterID(&cluster), "-", "", -1)

	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(id)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}
