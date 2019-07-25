package key

import (
	"fmt"
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

func PrivateRouteTableName(az string) string {
	return fmt.Sprintf("PrivateRouteTable-%s", az)
}

func PrivateSubnetName(az string) string {
	return fmt.Sprintf("PrivateSubnet-%s", az)
}

func PrivateSubnetRouteTableAssociationName(az string) string {
	return fmt.Sprintf("PrivateSubnetRouteTableAssociation-%s", az)
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
	return getter.GetLabels()[label.ReleaseVersion]
}

// SanitizeCFResourceName filters out all non-ascii alphanumberics from input
// string.
//
// Example: SanitizeCFResourceName("abc-123") == "abc123"
// Example2: SanitizeCFResourceName("Dear god why? щ（ﾟДﾟщ）") == "Deargodwhy"
//
func SanitizeCFResourceName(v string) string {
	var rs []rune

	// Start with true to capitalize first character.
	previousWasSkipped := true

	// Iterate over unicode characters and add numbers and ASCII letters title
	// cased.
	for _, r := range []rune(v) {
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

func StackNameTCNP(getter LabelsGetter) string {
	return fmt.Sprintf("cluster-%s-tcnp", getter.GetLabels()[label.Cluster])
}

func VPCPeeringRouteName(az string) string {
	return fmt.Sprintf("VPCPeeringRoute-%s", az)
}
