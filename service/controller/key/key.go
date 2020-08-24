package key

import (
	"crypto/sha1" // nolint:gosec
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v7/pkg/template"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/internal/templates/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/internal/templates/cloudformation/tccp"
)

const (
	// CloudProviderTagName is used to add Cloud Provider tags to AWS resources.
	CloudProviderTagName = "kubernetes.io/cluster/%s"

	// Cluster tag name for tagging all resources helping cost analysis in AWS.
	ClusterTagName = "giantswarm.io/cluster"

	// CloudProviderTagOwnedValue is used to indicate an AWS resource is owned
	// and managed by a cluster.
	CloudProviderTagOwnedValue = "owned"

	// ComponentOS is the name of the component specified in a Release CR which
	// determines the version of the OS to be used for tenant cluster nodes and
	// is ultimately transformed into an AMI based on TC region.
	ComponentOS = "containerlinux"

	// EnableTerminationProtection is used to protect the CF stacks from deletion.
	EnableTerminationProtection = true

	// InstallationTagName is used for AWS resource tagging.
	InstallationTagName = "giantswarm.io/installation"

	// OrganizationTagName is used for AWS resource tagging.
	OrganizationTagName = "giantswarm.io/organization"

	// ProfileNameTemplate will be included in the IAM instance profile name.
	ProfileNameTemplate = "EC2-K8S-Role"
	// RoleNameTemplate will be included in the IAM role name.
	RoleNameTemplate = "EC2-K8S-Role"
	// PolicyNameTemplate will be included in the IAM policy name.
	PolicyNameTemplate = "EC2-K8S-Policy"
	// LogDeliveryURI is used for setting the correct ACL in the access log bucket
	LogDeliveryURI = "uri=http://acs.amazonaws.com/groups/s3/LogDelivery"

	InstanceIDAnnotation = "aws-operator.giantswarm.io/instance"

	chinaAWSCliContainerRegistry   = "docker://registry-intl.cn-shanghai.aliyuncs.com/giantswarm/awscli:latest"
	defaultAWSCliContainerRegistry = "quay.io/coreos/awscli:025a357f05242fdad6a81e8a6b520098aa65a600"
	defaultDockerVolumeSizeGB      = "100"
)

const (
	DockerVolumeResourceNameKey   = "DockerVolumeResourceName"
	MasterIgnitionHashKey         = "MasterIgnitionHash"
	MasterImageIDKey              = "MasterImageID"
	MasterInstanceResourceNameKey = "MasterInstanceResourceName"
	MasterInstanceTypeKey         = "MasterInstanceType"
	MasterInstanceMonitoring      = "Monitoring"
	VersionBundleVersionKey       = "VersionBundleVersion"
	WorkerCountKey                = "WorkerCount"
	WorkerMaxKey                  = "WorkerMax"
	WorkerMinKey                  = "WorkerMin"
	WorkerDockerVolumeSizeKey     = "WorkerDockerVolumeSizeGB"
	WorkerIgnitionHashKey         = "WorkerIgnitionHash"
	WorkerImageIDKey              = "WorkerImageID"
	WorkerInstanceMonitoring      = "Monitoring"
	WorkerInstanceTypeKey         = "WorkerInstanceType"
)

const (
	ClusterIDLabel = "giantswarm.io/cluster"

	AnnotationEtcdDomain        = "giantswarm.io/etcd-domain"
	AnnotationPrometheusCluster = "giantswarm.io/prometheus-cluster"

	LabelApp           = "app"
	LabelCluster       = "giantswarm.io/cluster"
	LabelCustomer      = "customer"
	LabelOrganization  = "giantswarm.io/organization"
	LabelVersionBundle = "giantswarm.io/version-bundle"

	LegacyLabelCluster = "cluster"
)

const (
	TagCluster = "giantswarm.io/cluster"
)

const (
	NodeDrainerLifecycleHookName = "NodeDrainer"
	WorkerASGRef                 = "workerAutoScalingGroup"
)

const (
	KindMaster      = "master"
	KindIngress     = "ingress"
	KindWorker      = "worker"
	KindEtcd        = "etcd-elb"
	KindInternalAPI = "internal-api"
)

const (
	kubectlVersion                           = "1.16.4"
	kubernetesAPIHealthzVersion              = "0999549a4c334b646288d08bd2c781c6aae2e12f"
	kubernetesSetupNetworkEnvironmentVersion = "68e90113331feca3b9ffe6a75a601b381ba8c1f7"
)

func ClusterAPIEndpoint(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Kubernetes.API.Domain
}

func AutoScalingGroupName(customObject v1alpha1.AWSConfig, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func AvailabilityZone(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.AZ
}

func AWSCliContainerRegistry(customObject v1alpha1.AWSConfig) string {
	if IsChinaRegion(customObject) {
		return chinaAWSCliContainerRegistry
	}
	return defaultAWSCliContainerRegistry
}

func BucketName(customObject v1alpha1.AWSConfig, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", accountID, ClusterID(customObject))
}

// BucketObjectName computes the S3 object path to the actual cloud config (e.g. ignition/<hash>).
func BucketObjectName(hash string) string {
	return fmt.Sprintf("ignition/%s", hash)
}

func CredentialName(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.CredentialSecret.Name
}

func CredentialNamespace(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.CredentialSecret.Namespace
}

func CloudConfigSmallTemplates() []string {
	return []string{
		cloudconfig.Small,
	}
}

func CloudFormationGuestTemplates() []string {
	return []string{
		tccp.AutoScalingGroup,
		tccp.IAMPolicies,
		tccp.Instance,
		tccp.InternetGateway,
		tccp.LaunchConfiguration,
		tccp.LoadBalancers,
		tccp.Main,
		tccp.NatGateway,
		tccp.LifecycleHooks,
		tccp.Outputs,
		tccp.RecordSets,
		tccp.RouteTables,
		tccp.SecurityGroups,
		tccp.Subnets,
		tccp.VPC,
	}
}

func ClusterCloudProviderTag(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf(CloudProviderTagName, ClusterID(customObject))
}

func ClusterCustomer(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterEtcdDomain(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s:%d", customObject.Spec.Cluster.Etcd.Domain, customObject.Spec.Cluster.Etcd.Port)
}

func ClusterID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.ID
}

func ClusterNamespace(customObject v1alpha1.AWSConfig) string {
	return ClusterID(customObject)
}

// ClusterOrganization returns the org name from the custom object.
// It uses ClusterCustomer until this field is renamed in the custom object.
func ClusterOrganization(customObject v1alpha1.AWSConfig) string {
	return ClusterCustomer(customObject)
}

func ClusterTags(customObject v1alpha1.AWSConfig, installationName string) map[string]string {
	cloudProviderTag := ClusterCloudProviderTag(customObject)
	tags := map[string]string{
		cloudProviderTag:    CloudProviderTagOwnedValue,
		ClusterTagName:      ClusterID(customObject),
		InstallationTagName: installationName,
		OrganizationTagName: ClusterOrganization(customObject),
	}

	return tags
}

func CustomerID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func DefaultVersions() k8scloudconfig.Versions {
	return k8scloudconfig.Versions{
		KubernetesAPIHealthz:         kubernetesAPIHealthzVersion,
		KubernetesNetworkSetupDocker: kubernetesSetupNetworkEnvironmentVersion,
	}
}

func DockerVolumeResourceName(customObject v1alpha1.AWSConfig) string {
	return getResourcenameWithTimeHash("DockerVolume", customObject)
}

func DockerVolumeName(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-docker", ClusterID(customObject))
}

func EtcdVolumeName(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-etcd", ClusterID(customObject))
}

func LogVolumeName(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-log", ClusterID(customObject))
}

func EC2ServiceDomain(customObject v1alpha1.AWSConfig) string {
	domain := "ec2.amazonaws.com"

	if IsChinaRegion(customObject) {
		domain += ".cn"
	}

	return domain
}

func BaseDomain(customObject v1alpha1.AWSConfig) string {
	// TODO remove other zones and make it a BaseDomain in the CR.
	// CloudFormation creates a separate HostedZone with the same name.
	// Probably the easiest way for now is to just allow single domain for
	// everything which we do now.
	return customObject.Spec.AWS.HostedZones.API.Name
}

func HostedZoneNameAPI(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.HostedZones.API.Name
}

func HostedZoneNameEtcd(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.HostedZones.Etcd.Name
}

func HostedZoneNameIngress(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.HostedZones.Ingress.Name
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

func InternalLoadBalancerName(domainName string, cluster v1alpha1.AWSConfig) (string, error) {
	if ClusterID(cluster) == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "spec.cluster.id")
	}

	componentName, err := componentName(domainName)
	if err != nil {
		return "", microerror.Maskf(malformedCloudConfigKeyError, "spec.cluster.id")
	}

	lbName := fmt.Sprintf("%s-%s-internal", ClusterID(cluster), componentName)

	return lbName, nil
}

func IsChinaRegion(customObject v1alpha1.AWSConfig) bool {
	return strings.HasPrefix(Region(customObject), "cn-")
}

func IsDeleted(customObject v1alpha1.AWSConfig) bool {
	return customObject.GetDeletionTimestamp() != nil
}

func KubernetesAPISecurePort(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Kubernetes.API.SecurePort
}

func EtcdDomain(customObject v1alpha1.AWSConfig) string {
	return strings.Join([]string{"etcd", ClusterID(customObject), "k8s", BaseDomain(customObject)}, ".")
}

func EtcdPort(customObject v1alpha1.AWSConfig) int {
	return 2379
}

// LoadBalancerName produces a unique name for the load balancer.
// It takes the domain name, extracts the first subdomain, and combines it with the cluster name.
func LoadBalancerName(domainName string, cluster v1alpha1.AWSConfig) (string, error) {
	if ClusterID(cluster) == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "spec.cluster.id")
	}

	componentName, err := componentName(domainName)
	if err != nil {
		return "", microerror.Maskf(malformedCloudConfigKeyError, "spec.cluster.id")
	}

	lbName := fmt.Sprintf("%s-%s", ClusterID(cluster), componentName)

	return lbName, nil
}

func MachineDeploymentSubnet(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
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

func MasterInstanceResourceName(customObject v1alpha1.AWSConfig) string {
	return getResourcenameWithTimeHash("MasterInstance", customObject)
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

func MasterRoleARN(customObject v1alpha1.AWSConfig, accountID string) string {
	return baseRoleARN(customObject, accountID, "master")
}

func NATEIPName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "NATEIP"
	}
	return fmt.Sprintf("NATEIP%02d", idx)
}

func NATGatewayName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "NATGateway"
	}
	return fmt.Sprintf("NATGateway%02d", idx)
}

func NATRouteName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "NATRoute"
	}
	return fmt.Sprintf("NATRoute%02d", idx)
}

func PeerAccessRoleName(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(customObject))
}

func PolicyName(customObject v1alpha1.AWSConfig, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(customObject), profileType, PolicyNameTemplate)
}

func PrivateSubnetRouteTableAssociationName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PrivateSubnetRouteTableAssociation"
	}
	return fmt.Sprintf("PrivateSubnetRouteTableAssociation%02d", idx)
}

func PrivateRouteTableName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PrivateRouteTable"
	}
	return fmt.Sprintf("PrivateRouteTable%02d", idx)
}

func PrivateSubnetCIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.VPC.PrivateSubnetCIDR
}

func PrivateSubnetName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PrivateSubnet"
	}
	return fmt.Sprintf("PrivateSubnet%02d", idx)
}

func PublicSubnetCIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.VPC.PublicSubnetCIDR
}

func PublicSubnetRouteTableAssociationName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PublicSubnetRouteTableAssociation"
	}
	return fmt.Sprintf("PublicSubnetRouteTableAssociation%02d", idx)
}

func PublicRouteTableName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PublicRouteTable"
	}
	return fmt.Sprintf("PublicRouteTable%02d", idx)
}

func PublicSubnetName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PublicSubnet"
	}
	return fmt.Sprintf("PublicSubnet%02d", idx)
}

func CIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.VPC.CIDR
}

func Region(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.AWS.Region
}

func RegionARN(customObject v1alpha1.AWSConfig) string {
	regionARN := "aws"

	if IsChinaRegion(customObject) {
		regionARN += "-cn"
	}

	return regionARN
}

func RoleName(customObject v1alpha1.AWSConfig, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(customObject), profileType, RoleNameTemplate)
}

func RouteTableName(customObject v1alpha1.AWSConfig, suffix string, idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return fmt.Sprintf("%s-%s", ClusterID(customObject), suffix)
	}
	return fmt.Sprintf("%s-%s%02d", ClusterID(customObject), suffix, idx)
}

func S3ServiceDomain(customObject v1alpha1.AWSConfig) string {
	s3Domain := fmt.Sprintf("s3.%s.amazonaws.com", Region(customObject))

	if IsChinaRegion(customObject) {
		s3Domain += ".cn"
	}

	return s3Domain
}

func ScalingMax(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Scaling.Max
}

func ScalingMin(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Scaling.Min
}

func SecurityGroupName(customObject v1alpha1.AWSConfig, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func SmallCloudConfigPath(customObject v1alpha1.AWSConfig, accountID string, hash string) string {
	return fmt.Sprintf("%s/%s", BucketName(customObject, accountID), BucketObjectName(hash))
}

func SmallCloudConfigS3URL(customObject v1alpha1.AWSConfig, accountID string, hash string) string {
	return fmt.Sprintf("s3://%s", SmallCloudConfigPath(customObject, accountID, hash))
}

func SpecAvailabilityZones(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.AWS.AvailabilityZones
}

func StatusAvailabilityZones(customObject v1alpha1.AWSConfig) []v1alpha1.AWSConfigStatusAWSAvailabilityZone {
	return customObject.Status.AWS.AvailabilityZones
}

// StatusClusterNetworkCIDR returns the allocated tenant cluster subnet CIDR.
func StatusClusterNetworkCIDR(cluster v1alpha2.AWSCluster) string {
	return cluster.Status.Provider.Network.CIDR
}

// StatusAWSInfrastructureClusterNetworkCIDR returns the allocated tenant cluster subnet CIDR.
func StatusAWSInfrastructureClusterNetworkCIDR(cluster infrastructurev1alpha2.AWSCluster) string {
	return cluster.Status.Provider.Network.CIDR
}

// StatusNetworkCIDR returns the allocated tenant cluster subnet CIDR.
func StatusNetworkCIDR(customObject v1alpha1.AWSConfig) string {
	return customObject.Status.Cluster.Network.CIDR
}

func StatusScalingDesiredCapacity(customObject v1alpha1.AWSConfig) int {
	return customObject.Status.Cluster.Scaling.DesiredCapacity
}

func SubnetName(customObject v1alpha1.AWSConfig, suffix string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), suffix)
}

func TargetLogBucketName(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-g8s-access-logs", ClusterID(customObject))
}

func ToClusterEndpoint(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterAPIEndpoint(customObject), nil
}

func ToClusterID(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterID(customObject), nil
}

func ToClusterStatus(v interface{}) (v1alpha1.StatusCluster, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return customObject.Status.Cluster, nil
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

	customObject = *customObject.DeepCopy()

	return customObject, nil
}

func ToNodeCount(v interface{}) (int, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	// DesiredCapacity only returns number of workers, so we need to add the master
	nodeCount := customObject.Status.Cluster.Scaling.DesiredCapacity + 1

	return nodeCount, nil
}

func ToVersionBundleVersion(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return VersionBundleVersion(customObject), nil
}

// VersionBundleVersion returns the version contained in the Version Bundle.
func VersionBundleVersion(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.VersionBundle.Version
}

func VersionLabelSelector(enabled bool, overridenVersion string) string {
	if !enabled {
		return ""
	}

	version := project.Version()
	if overridenVersion != "" {
		version = overridenVersion
	}

	return label.OperatorVersion + "=" + version
}

func VPCPeeringRouteName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "VPCPeeringRoute"
	}
	return fmt.Sprintf("VPCPeeringRoute%02d", idx)
}

func WorkerCount(customObject v1alpha1.AWSConfig) int {
	return len(customObject.Spec.AWS.Workers)
}

// WorkerDockerVolumeSizeGB returns size of a docker volume configured for
// worker nodes. If there are no workers in custom object, 0 is returned as
// size.
func WorkerDockerVolumeSizeGB(customObject v1alpha1.AWSConfig) string {
	if len(customObject.Spec.AWS.Workers) <= 0 {
		return defaultDockerVolumeSizeGB
	}

	if customObject.Spec.AWS.Workers[0].DockerVolumeSizeGB <= 0 {
		return defaultDockerVolumeSizeGB
	}

	return strconv.Itoa(customObject.Spec.AWS.Workers[0].DockerVolumeSizeGB)
}

func WorkerInstanceType(customObject v1alpha1.AWSConfig) string {
	var instanceType string

	if len(customObject.Spec.AWS.Workers) > 0 {
		instanceType = customObject.Spec.AWS.Workers[0].InstanceType

	}

	return instanceType
}

func WorkerRoleARN(customObject v1alpha1.AWSConfig, accountID string) string {
	return baseRoleARN(customObject, accountID, "worker")
}

func baseRoleARN(customObject v1alpha1.AWSConfig, accountID string, kind string) string {
	clusterID := ClusterID(customObject)
	partition := RegionARN(customObject)

	return fmt.Sprintf("arn:%s:iam::%s:role/%s-%s-%s", partition, accountID, clusterID, kind, RoleNameTemplate)
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

// ImageID returns the EC2 AMI for the configured region and given version.
func ImageID(customObject v1alpha1.AWSConfig, release releasev1alpha1.Release) (string, error) {
	osVersion, err := OSVersion(release)
	if err != nil {
		return "", microerror.Mask(err)
	}

	regionAMIs, ok := amiInfo[osVersion]
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "no image id for version '%s'", osVersion)
	}

	region := Region(customObject)
	regionAMI, ok := regionAMIs[region]
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "no image id for region '%s'", region)
	}

	return regionAMI, nil
}

// getResourcenameWithTimeHash returns the string compared from specific prefix,
// time hash and cluster ID.
func getResourcenameWithTimeHash(prefix string, customObject v1alpha1.AWSConfig) string {
	clusterID := strings.Replace(ClusterID(customObject), "-", "", -1)

	h := sha1.New() // nolint:gosec
	_, _ = h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(clusterID)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}

func ComponentVersion(release releasev1alpha1.Release, componentName string) (string, error) {
	for _, component := range release.Spec.Components {
		if component.Name == componentName {
			return component.Version, nil
		}
	}
	return "", microerror.Maskf(notFoundError, "version for component %#v not found on release %#v", componentName, release.Name)
}

func OSVersion(release releasev1alpha1.Release) (string, error) {
	return ComponentVersion(release, ComponentOS)
}
