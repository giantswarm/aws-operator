package key

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/giantswarm/aws-operator/pkg/label"
)

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

func HealthCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}

func InternalELBNameAPI(getter LabelsGetter) string {
	return fmt.Sprintf("%s-api-internal", ClusterID(getter))
}

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}

func ImageID(region string) string {
	return imageIDs()[region]
}

func KubeletLabelsTCCPN(getter LabelsGetter) string {
	var labels string

	labels = ensureLabel(labels, label.Provider, "aws")
	labels = ensureLabel(labels, label.OperatorVersion, OperatorVersion(getter))

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

// S3ObjectPathTCCP computes the S3 object path to the cloud config uploaded
// for the TCCP stack.
//
//     version/3.4.0/cloudconfig/v_3_2_5/cluster-al9qy-tccp
//
func S3ObjectPathTCCP(getter LabelsGetter) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s", OperatorVersion(getter), CloudConfigVersion, StackNameTCCP(getter))
}

// S3ObjectPathTCCPN computes the S3 object path to the cloud config uploaded
// for the TCCPN stack. Note that the path is suffixed with the master ID, since
// Tenant Clusters may be Single Master or HA Masters, where the suffix -0
// indicates a Single Master configuration.
//
//     version/3.4.0/cloudconfig/v_3_2_5/cluster-al9qy-tccpn-2
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

func VPCPeeringRouteName(az string) string {
	return fmt.Sprintf("VPCPeeringRoute-%s", az)
}

// imageIDs returns our Container Linux AMIs for each active AWS region. Note
// that AMIs should always be for HVM virtualisation, not PV. Current Release is
// CoreOS Container Linux stable 2135.4.0. AMI IDs are copied from the following
// resource.
//
//     https://stable.release.core-os.net/amd64-usr/2135.4.0/coreos_production_ami_hvm.txt.
//
func imageIDs() map[string]string {
	return map[string]string{
		"ap-northeast-1": "ami-06443443a3ad575e0",
		"ap-northeast-2": "ami-05385569b790d035a",
		"ap-south-1":     "ami-05d7bc2359eaaecf1",
		"ap-southeast-1": "ami-0e69fd5ed05e58e4a",
		"ap-southeast-2": "ami-0af85d64c1d5aeae6",
		"ca-central-1":   "ami-00cbc28393f9da64c",
		"cn-north-1":     "ami-001272d09c87c54fa",
		"cn-northwest-1": "ami-0c08167b4fb0293c1",
		"eu-central-1":   "ami-038cea5071a5ee580",
		"eu-north-1":     "ami-01f28d71d1c924642",
		"eu-west-1":      "ami-067301c1a68e593f5",
		"eu-west-2":      "ami-0f5c4ede722171894",
		"eu-west-3":      "ami-07bf54c1c2b7c368e",
		"sa-east-1":      "ami-0d1ca6b44a76c404a",
		"us-east-1":      "ami-06d2804068b372d32",
		"us-east-2":      "ami-07ee0d30575e363c4",
		"us-gov-east-1":  "ami-0751c20ce4cb557df",
		"us-gov-west-1":  "ami-a9571fc8",
		"us-west-1":      "ami-0d05a67ab67139420",
		"us-west-2":      "ami-039eb9d6842534000",
	}
}

func isChinaRegion(region string) bool {
	return strings.HasPrefix(region, "cn-")
}
