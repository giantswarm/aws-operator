package key

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/giantswarm/aws-operator/pkg/label"
)

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

// BucketObjectName computes the S3 object path to the actual cloud config.
//
//     /version/3.4.0/cloudconfig/v_3_2_5/master
//     /version/3.4.0/cloudconfig/v_3_2_5/worker
//
func BucketObjectName(getter LabelsGetter, role string) string {
	return fmt.Sprintf("version/%s/cloudconfig/%s/%s", OperatorVersion(getter), CloudConfigVersion, role)
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

func ELBNameIngress(getter LabelsGetter) string {
	return fmt.Sprintf("%s-ingress", ClusterID(getter))
}

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}

func ImageID(region string) string {
	return imageIDs()[region]
}

func MachineDeploymentASGName(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tcnp-%s", ClusterID(getter), MachineDeploymentID(getter))
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

func RegionARN(region string) string {
	regionARN := "aws"

	if isChinaRegion(region) {
		regionARN += "-cn"
	}

	return regionARN
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.ReleaseVersion]
}

func RoleARNMaster(getter LabelsGetter, region string, accountID string) string {
	return baseRoleARN(getter, region, accountID, "master")
}

func RoleARNWorker(getter LabelsGetter, region string, accountID string) string {
	return baseRoleARN(getter, region, accountID, "worker")
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
	for _, r := range []rune(strings.Join(l, "-")) {
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

func SmallCloudConfigPath(getter LabelsGetter, accountID string, role string) string {
	return fmt.Sprintf("%s/%s", BucketName(getter, accountID), BucketObjectName(getter, role))
}

func SmallCloudConfigS3URL(getter LabelsGetter, accountID string, role string) string {
	return fmt.Sprintf("s3://%s", SmallCloudConfigPath(getter, accountID, role))
}

func StackNameCPF(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-cpf", ClusterID(getter))
}

func StackNameCPI(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-cpi", ClusterID(getter))
}

func StackNameTCCP(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tccp", ClusterID(getter))
}

func StackNameTCNP(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tcnp-%s", ClusterID(getter), MachineDeploymentID(getter))
}

func VPCPeeringRouteName(az string) string {
	return fmt.Sprintf("VPCPeeringRoute-%s", az)
}

func baseRoleARN(getter LabelsGetter, region string, accountID string, kind string) string {
	clusterID := ClusterID(getter)
	partition := RegionARN(region)

	return fmt.Sprintf("arn:%s:iam::%s:role/%s-%s-%s", partition, accountID, clusterID, kind, EC2RoleK8s)
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
		"ap-northeast-1": "ami-02e7b007b87514a38",
		"ap-northeast-2": "ami-0b5d1f638fb771cc9",
		"ap-south-1":     "ami-0db4916dd31b99465",
		"ap-southeast-1": "ami-01f2de2186e97c395",
		"ap-southeast-2": "ami-026d43721ef96eba8",
		"ca-central-1":   "ami-07d5bae9b2c4c9df1",
		"cn-north-1":     "ami-0dd65d250887524c1",
		"cn-northwest-1": "ami-0c63b500c3173c90e",
		"eu-central-1":   "ami-0eb0d9bb7ad1bd1e9",
		"eu-north-1":     "ami-0e3eca3c62f4c6311",
		"eu-west-1":      "ami-000307cf706ac9f94",
		"eu-west-2":      "ami-0322cee7ff4e446ce",
		"eu-west-3":      "ami-01c936a41649a8cda",
		"sa-east-1":      "ami-0b4101a238b99a929",
		"us-east-1":      "ami-00386353b49e325ba",
		"us-east-2":      "ami-064fe7e0332ae6407",
		"us-gov-east-1":  "ami-03e5a71feb2b7afd2",
		"us-gov-west-1":  "ami-272d6846",
		"us-west-1":      "ami-070bfb410b9f148c7",
		"us-west-2":      "ami-0a7e0ff8d31da1836",
	}
}

func isChinaRegion(region string) bool {
	return strings.HasPrefix(region, "cn-")
}
