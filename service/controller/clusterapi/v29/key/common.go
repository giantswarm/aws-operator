package key

import (
	"fmt"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
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
	var bs []byte
	for _, b := range []byte(v) {
		if ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z') || ('0' <= b && b <= '9') {
			bs = append(bs, b)
		}
	}
	return string(bs)
}

func VPCPeeringRouteName(az string) string {
	return fmt.Sprintf("VPCPeeringRoute-%s", az)
}
