package legacykey

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates/cloudformation/tccp"
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

	// RoleNameTemplate will be included in the IAM role name.
	RoleNameTemplate = "EC2-K8S-Role"
	// PolicyNameTemplate will be included in the IAM policy name.
	PolicyNameTemplate = "EC2-K8S-Policy"
	// LogDeliveryURI is used for setting the correct ACL in the access log bucket
	LogDeliveryURI = "uri=http://acs.amazonaws.com/groups/s3/LogDelivery"

	AnnotationInstanceID = "aws-operator.giantswarm.io/instance"

	defaultDockerVolumeSizeGB = "100"
)

const (
	MasterInstanceMonitoring = "Monitoring"
	WorkerCountKey           = "WorkerCount"
	WorkerMaxKey             = "WorkerMax"
	WorkerMinKey             = "WorkerMin"
	WorkerInstanceMonitoring = "Monitoring"
)

const (
	LabelApp           = "app"
	LabelCluster       = "giantswarm.io/cluster"
	LabelOrganization  = "giantswarm.io/organization"
	LabelVersionBundle = "giantswarm.io/version-bundle"
)

const (
	IngressControllerInsecurePort = 30010
	IngressControllerSecurePort   = 30011
)

const (
	KubernetesSecurePort = 443
)

const (
	KindMaster  = "master"
	KindIngress = "ingress"
	KindWorker  = "worker"
	KindEtcd    = "etcd-elb"
)

func ClusterAPIEndpoint(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Kubernetes.API.Domain
}

func BucketName(customObject v1alpha1.AWSConfig, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", accountID, ClusterID(customObject))
}

// BucketObjectName computes the S3 object path to the actual cloud config.
//
//     /version/3.4.0/cloudconfig/v_3_2_5/master
//     /version/3.4.0/cloudconfig/v_3_2_5/worker
//
func BucketObjectName(customObject v1alpha1.AWSConfig, role string) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s", ClusterVersion(customObject), CloudConfigVersion, role)
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

func ClusterEtcdEndpoint(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s:%d", customObject.Spec.Cluster.Etcd.Domain, customObject.Spec.Cluster.Etcd.Port)
}

func ClusterID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.ID
}

func ClusterNamespace(customObject v1alpha1.AWSConfig) string {
	return ClusterID(customObject)
}

func ClusterTags(customObject v1alpha1.AWSConfig, installationName string) map[string]string {
	cloudProviderTag := ClusterCloudProviderTag(customObject)
	tags := map[string]string{
		cloudProviderTag:    CloudProviderTagOwnedValue,
		ClusterTagName:      ClusterID(customObject),
		InstallationTagName: installationName,
		OrganizationTagName: OrganizationID(customObject),
	}

	return tags
}

func OrganizationID(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func DockerVolumeResourceName(customObject v1alpha1.AWSConfig) string {
	return getResourcenameWithTimeHash("DockerVolume", customObject)
}

func VolumeNameDocker(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-docker", ClusterID(customObject))
}

func VolumeNameEtcd(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-etcd", ClusterID(customObject))
}

func VolumeNameLog(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-log", ClusterID(customObject))
}

func EC2ServiceDomain(customObject v1alpha1.AWSConfig) string {
	domain := "ec2.amazonaws.com"

	if IsChinaRegion(customObject) {
		domain += ".cn"
	}

	return domain
}

func ClusterBaseDomain(customObject v1alpha1.AWSConfig) string {
	// TODO remove other zones and make it a ClusterBaseDomain in the CR.
	// CloudFormation creates a separate HostedZone with the same name.
	// Probably the easiest way for now is to just allow single domain for
	// everything which we do now.
	return customObject.Spec.AWS.HostedZones.API.Name
}

func IsChinaRegion(customObject v1alpha1.AWSConfig) bool {
	return strings.HasPrefix(Region(customObject), "cn-")
}

func EtcdDomain(customObject v1alpha1.AWSConfig) string {
	return strings.Join([]string{"etcd", ClusterID(customObject), "k8s", ClusterBaseDomain(customObject)}, ".")
}

func EtcdPort(customObject v1alpha1.AWSConfig) int {
	return 2379
}

func ELBNameAPI(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-api", ClusterID(customObject))
}

func ELBNameEtcd(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-etcd", ClusterID(customObject))
}

func ELBNameIngress(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-ingress", ClusterID(customObject))
}

func StackNameCPF(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("cluster-%s-host-main", ClusterID(customObject))
}

func StackNameCPI(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("cluster-%s-host-setup", ClusterID(customObject))
}

func StackNameTCCP(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("cluster-%s-guest-main", ClusterID(customObject))
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

func RoleARNMaster(customObject v1alpha1.AWSConfig, accountID string) string {
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

func RolePeerAccess(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-vpc-peer-access", ClusterID(customObject))
}

func PolicyNameMaster(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-master-%s", ClusterID(customObject), PolicyNameTemplate)
}

func PolicyNameWorker(customObject v1alpha1.AWSConfig) string {
	return fmt.Sprintf("%s-worker-%s", ClusterID(customObject), PolicyNameTemplate)
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

func PrivateSubnetName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PrivateSubnet"
	}
	return fmt.Sprintf("PrivateSubnet%02d", idx)
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

func ProfileName(customObject v1alpha1.AWSConfig, profileType string) string {
	return RoleName(customObject, profileType)
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

func WorkerScalingMax(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Scaling.Max
}

func WorkerScalingMin(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.Cluster.Scaling.Min
}

func SecurityGroupName(customObject v1alpha1.AWSConfig, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(customObject), groupName)
}

func SmallCloudConfigPath(customObject v1alpha1.AWSConfig, accountID string, role string) string {
	return fmt.Sprintf("%s/%s", BucketName(customObject, accountID), BucketObjectName(customObject, role))
}

func SmallCloudConfigS3HTTPURL(customObject v1alpha1.AWSConfig, accountID string, role string) string {
	return fmt.Sprintf("https://%s/%s", S3ServiceDomain(customObject), SmallCloudConfigPath(customObject, accountID, role))
}

func SmallCloudConfigS3URL(customObject v1alpha1.AWSConfig, accountID string, role string) string {
	return fmt.Sprintf("s3://%s", SmallCloudConfigPath(customObject, accountID, role))
}

func SpecAvailabilityZones(customObject v1alpha1.AWSConfig) int {
	return customObject.Spec.AWS.AvailabilityZones
}

func StatusAvailabilityZones(customObject v1alpha1.AWSConfig) []v1alpha1.AWSConfigStatusAWSAvailabilityZone {
	return customObject.Status.AWS.AvailabilityZones
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

	return ClusterVersion(customObject), nil
}

func ClusterVersion(customObject v1alpha1.AWSConfig) string {
	return customObject.Spec.VersionBundle.Version
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

func RoleARNWorker(customObject v1alpha1.AWSConfig, accountID string) string {
	return baseRoleARN(customObject, accountID, "worker")
}

func baseRoleARN(customObject v1alpha1.AWSConfig, accountID string, kind string) string {
	clusterID := ClusterID(customObject)
	partition := RegionARN(customObject)

	return fmt.Sprintf("arn:%s:iam::%s:role/%s-%s-%s", partition, accountID, clusterID, kind, RoleNameTemplate)
}

// ImageID returns the EC2 AMI for the configured region.
func ImageID(customObject v1alpha1.AWSConfig) string {
	region := Region(customObject)

	/*
		Container Linux AMIs for each active AWS region.

		NOTE 1: AMIs should always be for HVM virtualisation and not PV.
		NOTE 2: You also need to update the tests.

		service/controller/v27/key/key_test.go
		service/controller/v27/adapter/adapter_test.go
		service/controller/v27/resource/cloudformation/main_stack_test.go

		Current Release: CoreOS Container Linux stable 2023.5.0 (HVM)
		AMI IDs copied from https://stable.release.core-os.net/amd64-usr/2023.5.0/coreos_production_ami_hvm.txt.
	*/
	imageIDs := map[string]string{
		"ap-northeast-1": "ami-0d3a9785820124591",
		"ap-northeast-2": "ami-03230b2fa6af112bf",
		"ap-south-1":     "ami-0b85fd1356963d2ee",
		"ap-southeast-1": "ami-0f8a9aa9857d8af7e",
		"ap-southeast-2": "ami-0e87752a1d331823a",
		"ca-central-1":   "ami-0c0100bac23bb1d39",
		"cn-north-1":     "ami-01e99c7e0a343d325",
		"cn-northwest-1": "ami-0773341917796083a",
		"eu-central-1":   "ami-012abdf0d2781f0a5",
		"eu-north-1":     "ami-09fbda19ac2fc6c3f",
		"eu-west-1":      "ami-01f5fbceb7a9fa4d0",
		"eu-west-2":      "ami-069966bea0809e21d",
		"eu-west-3":      "ami-0194c504244182155",
		"sa-east-1":      "ami-0cd830cc037613a7d",
		"us-east-1":      "ami-08e58b93705fb503f",
		"us-east-2":      "ami-03172282aaa2899be",
		"us-gov-east-1":  "ami-0ff9e298ea0bacf53",
		"us-gov-west-1":  "ami-e7f59e86",
		"us-west-1":      "ami-08d3e245ebf4d560f",
		"us-west-2":      "ami-0a4f49b2488e15346",
	}

	return imageIDs[region]
}

// getResourcenameWithTimeHash returns the string compared from specific prefix,
// time hash and cluster ID.
func getResourcenameWithTimeHash(prefix string, customObject v1alpha1.AWSConfig) string {
	clusterID := strings.Replace(ClusterID(customObject), "-", "", -1)

	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(clusterID)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}
