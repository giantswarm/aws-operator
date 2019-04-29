package key

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/templates/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/templates/cloudformation/tccp"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use.
	// It is used in the main stack output and S3 object paths.
	CloudConfigVersion = "v_4_0_0"

	// CloudProviderTagName is used to add Cloud Provider tags to AWS resources.
	CloudProviderTagName = "kubernetes.io/cluster/%s"

	// Cluster tag name for tagging all resources helping cost analysis in AWS.
	ClusterTagName = "giantswarm.io/cluster"

	// CloudProviderTagOwnedValue is used to indicate an AWS resource is owned
	// and managed by a cluster.
	CloudProviderTagOwnedValue = "owned"

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
	MasterImageIDKey              = "MasterImageID"
	MasterInstanceResourceNameKey = "MasterInstanceResourceName"
	MasterInstanceTypeKey         = "MasterInstanceType"
	MasterInstanceMonitoring      = "Monitoring"
	MasterCloudConfigVersionKey   = "MasterCloudConfigVersion"
	VersionBundleVersionKey       = "VersionBundleVersion"
	WorkerCountKey                = "WorkerCount"
	WorkerMaxKey                  = "WorkerMax"
	WorkerMinKey                  = "WorkerMin"
	WorkerDockerVolumeSizeKey     = "WorkerDockerVolumeSizeGB"
	WorkerImageIDKey              = "WorkerImageID"
	WorkerInstanceMonitoring      = "Monitoring"
	WorkerInstanceTypeKey         = "WorkerInstanceType"
	WorkerCloudConfigVersionKey   = "WorkerCloudConfigVersion"
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
	NodeDrainerLifecycleHookName = "NodeDrainer"
	WorkerASGRef                 = "workerAutoScalingGroup"
)

const (
	KindMaster  = "master"
	KindIngress = "ingress"
	KindWorker  = "worker"
	KindEtcd    = "etcd-elb"
)

func ClusterAPIEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("api.%s.%s", ClusterID(cluster), BaseDomain(cluster))
}

func AutoScalingGroupName(cluster v1alpha1.Cluster, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(cluster), groupName)
}

func AWSCliContainerRegistry(cluster v1alpha1.Cluster) string {
	if IsChinaRegion(cluster) {
		return chinaAWSCliContainerRegistry
	}
	return defaultAWSCliContainerRegistry
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

func CredentialName(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Provider.CredentialSecret.Name
}

func CredentialNamespace(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Provider.CredentialSecret.Namespace
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

func ClusterCloudProviderTag(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf(CloudProviderTagName, ClusterID(cluster))
}

func ClusterEtcdDomain(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("etcd.%s.%s", ClusterID(cluster), providerSpec(cluster).Cluster.DNS.Domain)
}

func ClusterID(cluster v1alpha1.Cluster) string {
	return providerStatus(cluster).Cluster.ID
}

func ClusterNamespace(cluster v1alpha1.Cluster) string {
	return ClusterID(cluster)
}

func ClusterTags(cluster v1alpha1.Cluster, installationName string) map[string]string {
	cloudProviderTag := ClusterCloudProviderTag(cluster)
	tags := map[string]string{
		cloudProviderTag:    CloudProviderTagOwnedValue,
		ClusterTagName:      ClusterID(cluster),
		InstallationTagName: installationName,
		OrganizationTagName: OrganizationID(cluster),
	}

	return tags
}

func OrganizationID(cluster v1alpha1.Cluster) string {
	return cluster.Spec.Cluster.Customer.ID
}

func DockerVolumeResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("DockerVolume", cluster)
}

func DockerVolumeName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-docker", ClusterID(cluster))
}

func EtcdVolumeName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-etcd", ClusterID(cluster))
}

func LogVolumeName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-log", ClusterID(cluster))
}

func EC2ServiceDomain(cluster v1alpha1.Cluster) string {
	domain := "ec2.amazonaws.com"

	if IsChinaRegion(cluster) {
		domain += ".cn"
	}

	return domain
}

func BaseDomain(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Cluster.DNS.Domain
}

func InstanceProfileName(cluster v1alpha1.Cluster, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(cluster), profileType, ProfileNameTemplate)
}

func IsChinaRegion(cluster v1alpha1.Cluster) bool {
	return strings.HasPrefix(Region(cluster), "cn-")
}

func IsDeleted(cluster v1alpha1.Cluster) bool {
	return cluster.GetDeletionTimestamp() != nil
}

func KubernetesAPISecurePort(cluster v1alpha1.Cluster) int {
	return cluster.Spec.Cluster.Kubernetes.API.SecurePort
}

func EtcdDomain(cluster v1alpha1.Cluster) string {
	return strings.Join([]string{"etcd", ClusterID(cluster), "k8s", BaseDomain(cluster)}, ".")
}

func EtcdPort(cluster v1alpha1.Cluster) int {
	return 2379
}

// LoadBalancerName produces a unique name for the load balancer.
// It takes the domain name, extracts the first subdomain, and combines it with the cluster name.
func LoadBalancerName(domainName string, cluster v1alpha1.Cluster) (string, error) {
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

func MainGuestStackName(cluster v1alpha1.Cluster) string {
	clusterID := ClusterID(cluster)

	return fmt.Sprintf("cluster-%s-guest-main", clusterID)
}

func MainHostPreStackName(cluster v1alpha1.Cluster) string {
	clusterID := ClusterID(cluster)

	return fmt.Sprintf("cluster-%s-host-setup", clusterID)
}

func MainHostPostStackName(cluster v1alpha1.Cluster) string {
	clusterID := ClusterID(cluster)

	return fmt.Sprintf("cluster-%s-host-main", clusterID)
}

func MasterCount(cluster v1alpha1.Cluster) int {
	return len(cluster.Spec.AWS.Masters)
}

func MasterImageID(cluster v1alpha1.Cluster) string {
	var imageID string

	if len(cluster.Spec.AWS.Masters) > 0 {
		imageID = cluster.Spec.AWS.Masters[0].ImageID
	}

	return imageID
}

func MasterInstanceResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("MasterInstance", cluster)
}

func MasterInstanceName(cluster v1alpha1.Cluster) string {
	clusterID := ClusterID(cluster)

	return fmt.Sprintf("%s-master", clusterID)
}

func MasterInstanceType(cluster v1alpha1.Cluster) string {
	var instanceType string

	if len(cluster.Spec.AWS.Masters) > 0 {
		instanceType = cluster.Spec.AWS.Masters[0].InstanceType
	}

	return instanceType
}

func MasterRoleARN(cluster v1alpha1.Cluster, accountID string) string {
	return baseRoleARN(cluster, accountID, "master")
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

func PeerAccessRoleName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(cluster))
}

func PeerID(cluster v1alpha1.Cluster) string {
	return cluster.Spec.AWS.VPC.PeerID
}

func PolicyName(cluster v1alpha1.Cluster, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(cluster), profileType, PolicyNameTemplate)
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

func PrivateSubnetCIDR(cluster v1alpha1.Cluster) string {
	return cluster.Spec.AWS.VPC.PrivateSubnetCIDR
}

func PrivateSubnetName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PrivateSubnet"
	}
	return fmt.Sprintf("PrivateSubnet%02d", idx)
}

func PublicSubnetCIDR(cluster v1alpha1.Cluster) string {
	return cluster.Spec.AWS.VPC.PublicSubnetCIDR
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

func CIDR(cluster v1alpha1.Cluster) string {
	return cluster.Spec.AWS.VPC.CIDR
}

func Region(cluster v1alpha1.Cluster) string {
	return cluster.Spec.AWS.Region
}

func RegionARN(cluster v1alpha1.Cluster) string {
	regionARN := "aws"

	if IsChinaRegion(cluster) {
		regionARN += "-cn"
	}

	return regionARN
}

func RoleName(cluster v1alpha1.Cluster, profileType string) string {
	return fmt.Sprintf("%s-%s-%s", ClusterID(cluster), profileType, RoleNameTemplate)
}

func RouteTableName(cluster v1alpha1.Cluster, suffix string, idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return fmt.Sprintf("%s-%s", ClusterID(cluster), suffix)
	}
	return fmt.Sprintf("%s-%s%02d", ClusterID(cluster), suffix, idx)
}

func S3ServiceDomain(cluster v1alpha1.Cluster) string {
	s3Domain := fmt.Sprintf("s3.%s.amazonaws.com", Region(cluster))

	if IsChinaRegion(cluster) {
		s3Domain += ".cn"
	}

	return s3Domain
}

func ScalingMax(cluster v1alpha1.Cluster) int {
	return cluster.Spec.Cluster.Scaling.Max
}

func ScalingMin(cluster v1alpha1.Cluster) int {
	return cluster.Spec.Cluster.Scaling.Min
}

func SecurityGroupName(cluster v1alpha1.Cluster, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(cluster), groupName)
}

func SmallCloudConfigPath(cluster v1alpha1.Cluster, accountID string, role string) string {
	return fmt.Sprintf("%s/%s", BucketName(cluster, accountID), BucketObjectName(cluster, role))
}

func SmallCloudConfigS3HTTPURL(cluster v1alpha1.Cluster, accountID string, role string) string {
	return fmt.Sprintf("https://%s/%s", S3ServiceDomain(cluster), SmallCloudConfigPath(cluster, accountID, role))
}

func SmallCloudConfigS3URL(cluster v1alpha1.Cluster, accountID string, role string) string {
	return fmt.Sprintf("s3://%s", SmallCloudConfigPath(cluster, accountID, role))
}

func SpecAvailabilityZones(cluster v1alpha1.Cluster) int {
	return cluster.Spec.AWS.AvailabilityZones
}

func StatusAvailabilityZones(cluster v1alpha1.Cluster) []v1alpha1.AWSConfigStatusAWSAvailabilityZone {
	return cluster.Status.AWS.AvailabilityZones
}

// StatusNetworkCIDR returns the allocated tenant cluster subnet CIDR.
func StatusNetworkCIDR(cluster v1alpha1.Cluster) string {
	return cluster.Status.Cluster.Network.CIDR
}

func StatusScalingDesiredCapacity(cluster v1alpha1.Cluster) int {
	return cluster.Status.Cluster.Scaling.DesiredCapacity
}

func SubnetName(cluster v1alpha1.Cluster, suffix string) string {
	return fmt.Sprintf("%s-%s", ClusterID(cluster), suffix)
}

func TargetLogBucketName(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-g8s-access-logs", ClusterID(cluster))
}

func ToClusterEndpoint(v interface{}) (string, error) {
	cluster, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterAPIEndpoint(cluster), nil
}

func ToClusterID(v interface{}) (string, error) {
	cluster, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterID(cluster), nil
}

func ToClusterStatus(v interface{}) (v1alpha1.StatusCluster, error) {
	cluster, err := ToCustomObject(v)
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return cluster.Status.Cluster, nil
}

func ToCustomObject(v interface{}) (v1alpha1.Cluster, error) {
	if v == nil {
		return v1alpha1.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Cluster{}, v)
	}

	customObjectPointer, ok := v.(*v1alpha1.Cluster)
	if !ok {
		return v1alpha1.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Cluster{}, v)
	}
	cluster := *customObjectPointer

	cluster = *cluster.DeepCopy()

	return cluster, nil
}

func ToNodeCount(v interface{}) (int, error) {
	cluster, err := ToCustomObject(v)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	nodeCount := MasterCount(cluster) + WorkerCount(cluster)

	return nodeCount, nil
}

func ToVersionBundleVersion(v interface{}) (string, error) {
	cluster, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return VersionBundleVersion(cluster), nil
}

// VersionBundleVersion returns the version contained in the Version Bundle.
func VersionBundleVersion(cluster v1alpha1.Cluster) string {
	return cluster.Spec.VersionBundle.Version
}

func VPCPeeringRouteName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "VPCPeeringRoute"
	}
	return fmt.Sprintf("VPCPeeringRoute%02d", idx)
}

func WorkerCount(cluster v1alpha1.Cluster) int {
	return len(cluster.Spec.AWS.Workers)
}

// WorkerDockerVolumeSizeGB returns size of a docker volume configured for
// worker nodes. If there are no workers in custom object, 0 is returned as
// size.
func WorkerDockerVolumeSizeGB(cluster v1alpha1.Cluster) string {
	if len(cluster.Spec.AWS.Workers) <= 0 {
		return defaultDockerVolumeSizeGB
	}

	if cluster.Spec.AWS.Workers[0].DockerVolumeSizeGB <= 0 {
		return defaultDockerVolumeSizeGB
	}

	return strconv.Itoa(cluster.Spec.AWS.Workers[0].DockerVolumeSizeGB)
}

func WorkerImageID(cluster v1alpha1.Cluster) string {
	var imageID string

	if len(cluster.Spec.AWS.Workers) > 0 {
		imageID = cluster.Spec.AWS.Workers[0].ImageID
	}

	return imageID
}

func WorkerInstanceType(cluster v1alpha1.Cluster) string {
	var instanceType string

	if len(cluster.Spec.AWS.Workers) > 0 {
		instanceType = cluster.Spec.AWS.Workers[0].InstanceType

	}

	return instanceType
}

func WorkerRoleARN(cluster v1alpha1.Cluster, accountID string) string {
	return baseRoleARN(cluster, accountID, "worker")
}

func baseRoleARN(cluster v1alpha1.Cluster, accountID string, kind string) string {
	clusterID := ClusterID(cluster)
	partition := RegionARN(cluster)

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

// ImageID returns the EC2 AMI for the configured region.
func ImageID(cluster v1alpha1.Cluster) (string, error) {
	region := Region(cluster)

	/*
		Container Linux AMIs for each active AWS region.

		NOTE 1: AMIs should always be for HVM virtualisation and not PV.
		NOTE 2: You also need to update the tests.

		service/controller/v26/key/key_test.go
		service/controller/v26/adapter/adapter_test.go
		service/controller/v26/resource/cloudformation/main_stack_test.go

		Current Release: CoreOS Container Linux stable 2023.4.0 (HVM)
		AMI IDs copied from https://stable.release.core-os.net/amd64-usr/2023.4.0/coreos_production_ami_hvm.txt.
	*/
	imageIDs := map[string]string{
		"ap-northeast-1": "ami-003b3a37a48d799cf",
		"ap-northeast-2": "ami-0c2d3bd39b13c3b2d",
		"ap-south-1":     "ami-0bd5eb3e67407e0df",
		"ap-southeast-1": "ami-07aafbd1f2a182cd4",
		"ap-southeast-2": "ami-0cb589c5f6134f078",
		"ca-central-1":   "ami-0952a9471ff71919e",
		"cn-north-1":     "ami-0caaf17a3032c1b56",
		"cn-northwest-1": "ami-0a863f3b0a0720e6a",
		"eu-central-1":   "ami-015e6cb33a709348e",
		"eu-west-1":      "ami-04d747d892ccd652a",
		"eu-west-2":      "ami-056a316ba69c9d9e8",
		"eu-west-3":      "ami-026d41122f47f745e",
		"sa-east-1":      "ami-0e9521088a80c2a02",
		"us-east-1":      "ami-09d5d3bcd3e0e5c30",
		"us-east-2":      "ami-02accfa372062664b",
		"us-gov-west-1":  "ami-07600866",
		"us-west-1":      "ami-0481a60675f6ea007",
		"us-west-2":      "ami-025acbb0fb1db6a27",
	}

	imageID, ok := imageIDs[region]
	if !ok {
		return "", microerror.Maskf(invalidConfigError, "no image id for region '%s'", region)
	}

	return imageID, nil
}

// getResourcenameWithTimeHash returns the string compared from specific prefix,
// time hash and cluster ID.
func getResourcenameWithTimeHash(prefix string, cluster v1alpha1.Cluster) string {
	clusterID := strings.Replace(ClusterID(cluster), "-", "", -1)

	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(clusterID)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}
