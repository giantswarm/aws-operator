package key

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
)

const (
	AWSCNIComponentName          = "aws-cni"
	AWSCNIDefaultMinimumIPTarget = "40"
	AWSCNIDefaultWarmIPTarget    = "10"

	ELBInstanceStateInService = "InService"

	DrainerResyncPeriod = time.Minute * 2

	DefaultPauseTimeBetweenUpdates = "PT15M"
)

const (
	// TerminateUnhealthyNodeResyncPeriod defines resync period for the terminateunhealthynode controller
	TerminateUnhealthyNodeResyncPeriod = time.Minute * 3
)

// AMI returns the EC2 AMI for the configured region and given version.
func AMI(region string, release releasev1alpha1.Release) (string, error) {
	osVersion, err := OSVersion(release)
	if err != nil {
		return "", microerror.Mask(err)
	}

	regionAMIs, ok := amiInfo[osVersion]
	if !ok {
		return "", microerror.Maskf(notFoundError, "no image id for version '%s'", osVersion)
	}

	regionAMI, ok := regionAMIs[region]
	if !ok {
		return "", microerror.Maskf(notFoundError, "no image id for region '%s'", region)
	}

	return regionAMI, nil
}

func AWSCNINATRouteName(az string) string {
	return fmt.Sprintf("AWSCNINATRoute-%s", az)
}

func AWSCNIRouteTableName(az string) string {
	return fmt.Sprintf("AWSCNIRouteTable-%s", az)
}

func AWSCNISubnetName(az string) string {
	return fmt.Sprintf("AWSCNISubnet-%s", az)
}

func AWSCNISubnetRouteTableAssociationName(az string) string {
	return fmt.Sprintf("AWSCNISubnetRouteTableAssociation-%s", az)
}

func AWSTags(getter LabelsGetter, installationName string) map[string]string {
	TagCloudProvider := ClusterCloudProviderTag(getter)

	tags := map[string]string{
		TagCloudProvider: "owned",
		TagCluster:       ClusterID(getter),
		TagInstallation:  installationName,
		TagOrganization:  OrganizationID(getter),
	}

	return tags
}

// AvailabilityZoneRegionSuffix takes region's full name and returns its
// suffix: e.g. "eu-central-1b" -> "1b"
func AvailabilityZoneRegionSuffix(az string) string {
	elements := strings.Split(az, "-")
	return elements[len(elements)-1]
}

func BucketName(getter LabelsGetter, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s", accountID, ClusterID(getter))
}

func ClusterCloudProviderTag(getter LabelsGetter) string {
	return fmt.Sprintf("kubernetes.io/cluster/%s", ClusterID(getter))
}

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func EC2ServiceDomain(region string) string {
	domain := "ec2.amazonaws.com"

	if isChinaRegion(region) {
		domain += ".cn"
	}

	return domain
}

func ELBNameAPI(getter LabelsGetter) string {
	return fmt.Sprintf("%s-api", ClusterID(getter))
}

func ELBNameEtcd(getter LabelsGetter) string {
	return fmt.Sprintf("%s-etcd", ClusterID(getter))
}

func HealthCheckTCPTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}

func HealthCheckHTTPTarget(port int) string {
	return fmt.Sprintf("HTTP:%d/healthz", port)
}

func InternalELBNameAPI(getter LabelsGetter) string {
	return fmt.Sprintf("%s-api-internal", ClusterID(getter))
}

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}

func KubeletLabelsTCCPN(getter LabelsGetter, masterID int) string {
	var labels string

	labels = ensureLabel(labels, label.Provider, "aws")
	labels = ensureLabel(labels, label.OperatorVersion, OperatorVersion(getter))
	labels = ensureLabel(labels, label.ControlPlane, ControlPlaneID(getter))
	labels = ensureLabel(labels, label.MasterID, fmt.Sprintf("%d", masterID))

	return labels
}

func KubeletLabelsTCNP(getter LabelsGetter) string {
	var labels string

	labels = ensureLabel(labels, label.Provider, "aws")
	labels = ensureLabel(labels, label.OperatorVersion, OperatorVersion(getter))
	labels = ensureLabel(labels, label.MachineDeployment, MachineDeploymentID(getter))

	return labels
}

func MachineDeploymentID(getter LabelsGetter) string {
	return getter.GetLabels()[label.MachineDeployment]
}

func NATEIPName(az string) string {
	return fmt.Sprintf("NATEIP-%s", az)
}

func NATGatewayName(az string) string {
	return fmt.Sprintf("NATGateway-%s", az)
}

func NATRouteName(az string) string {
	return fmt.Sprintf("NATRoute-%s", az)
}

func OperatorVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.OperatorVersion]
}

func OrganizationID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Organization]
}

func PrivateInternetGatewayRouteName(az string) string {
	return fmt.Sprintf("PrivateInternetGatewayRoute-%s", az)
}

func PrivateRouteTableName(az string) string {
	return fmt.Sprintf("PrivateRouteTable-%s", az)
}

func PrivateSubnetName(az string) string {
	return fmt.Sprintf("PrivateSubnet-%s", az)
}

func PrivateSubnetRouteTableAssociationName(az string) string {
	return fmt.Sprintf("PrivateSubnetRouteTableAssociation-%s", az)
}

func PublicInternetGatewayRouteName(az string) string {
	return fmt.Sprintf("PublicInternetGatewayRoute-%s", az)
}

func PublicSubnetName(az string) string {
	return fmt.Sprintf("PublicSubnet-%s", az)
}

func PublicRouteTableName(az string) string {
	return fmt.Sprintf("PublicRouteTable-%s", az)
}

func PublicSubnetRouteTableAssociationName(az string) string {
	return fmt.Sprintf("PublicSubnetRouteTableAssociation-%s", az)
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.Release]
}

func ReleaseName(releaseVersion string) string {
	return fmt.Sprintf("v%s", releaseVersion)
}

func RegionARN(region string) string {
	regionARN := "aws"

	if isChinaRegion(region) {
		regionARN += "-cn"
	}

	return regionARN
}

func RoleARNMaster(getter LabelsGetter, region string, accountID string) string {
	clusterID := ClusterID(getter)
	partition := RegionARN(region)

	return fmt.Sprintf("arn:%s:iam::%s:role/%s-master-%s", partition, accountID, clusterID, EC2RoleK8s)
}

func RoleARNWorker(getter LabelsGetter, region string, accountID string) string {
	clusterID := ClusterID(getter)
	partition := RegionARN(region)

	return fmt.Sprintf("arn:%s:iam::%s:role/gs-cluster-%s-role-*", partition, accountID, clusterID)
}

// S3ObjectPathTCCPN computes the S3 object path to the cloud config uploaded
// for the TCCPN stack. Note that the path is suffixed with the master ID, since
// Tenant Clusters may be Single Master or HA Masters, where the suffix -0
// indicates a Single Master configuration.
//
//     version/3.4.0/cloudconfig/v_3_2_5/cluster-al9qy-tccpn-a2wax-2
//
func S3ObjectPathTCCPN(cr LabelsGetter, id int) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s-%d", OperatorVersion(cr), CloudConfigVersion, StackNameTCCPN(cr), id)
}

// S3ObjectPathTCNP computes the S3 object path to the cloud config uploaded for
// the TCCP stack.
//
//     version/3.4.0/cloudconfig/v_3_2_5/cluster-al9qy-tcnp-g3j50
//
func S3ObjectPathTCNP(getter LabelsGetter) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s", OperatorVersion(getter), CloudConfigVersion, StackNameTCNP(getter))
}

// SanitizeCFResourceName filters out all non-ascii alphanumberics from input
// string.
//
//     SanitizeCFResourceName("abc-123") == "abc123"
//     SanitizeCFResourceName("abc", "123") == "abc123"
//     SanitizeCFResourceName("Dear god why? щ（ﾟДﾟщ）") == "Deargodwhy"
//
func SanitizeCFResourceName(l ...string) string {
	var rs []rune

	// Start with true to capitalize first character.
	previousWasSkipped := true

	// Iterate over unicode characters and add numbers and ASCII letters title
	// cased.
	for _, r := range strings.Join(l, "-") {
		if unicode.IsDigit(r) || (unicode.IsLetter(r) && utf8.RuneLen(r) == 1) {
			if previousWasSkipped {
				rs = append(rs, unicode.ToTitle(r))
			} else {
				rs = append(rs, r)
			}
			previousWasSkipped = false
		} else {
			previousWasSkipped = true
		}
	}

	return string(rs)
}

func SecurityGroupName(getter LabelsGetter, groupName string) string {
	return fmt.Sprintf("%s-%s", ClusterID(getter), groupName)
}

func StackComplete(status string) bool {
	return strings.Contains(status, "COMPLETE")
}

func StackInProgress(status string) bool {
	return strings.Contains(status, "IN_PROGRESS")
}

func StackNameTCCP(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tccp", ClusterID(getter))
}

func StackNameTCCPF(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tccpf", ClusterID(getter))
}

func StackNameTCCPI(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tccpi", ClusterID(getter))
}

func StackNameTCCPN(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tccpn", ClusterID(getter))
}

func StackNameTCNP(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tcnp-%s", ClusterID(getter), MachineDeploymentID(getter))
}

func StackNameTCNPF(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tcnpf-%s", ClusterID(getter), MachineDeploymentID(getter))
}

func TargetLogBucketName(getter LabelsGetter, accountID string) string {
	return fmt.Sprintf("%s-g8s-%s-access-logs", accountID, ClusterID(getter))
}

func VPCPeeringRouteName(az string) string {
	return fmt.Sprintf("VPCPeeringRoute-%s", az)
}

func isChinaRegion(region string) bool {
	return strings.HasPrefix(region, "cn-")
}

func IsWrongDrainerConfig(dc *g8sv1alpha1.DrainerConfig, clusterID string, instanceId string) bool {
	return dc.Labels[TagCluster] != clusterID || dc.Annotations[annotation.InstanceID] != instanceId
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
