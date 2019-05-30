package key

import "fmt"

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
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

func PrivateSubnetRouteTableAssociationName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PrivateSubnetRouteTableAssociation"
	}
	return fmt.Sprintf("PrivateSubnetRouteTableAssociation%02d", idx)
}

func PublicSubnetName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PublicSubnet"
	}
	return fmt.Sprintf("PublicSubnet%02d", idx)
}

func PublicRouteTableName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PublicRouteTable"
	}
	return fmt.Sprintf("PublicRouteTable%02d", idx)
}

func PublicSubnetRouteTableAssociationName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "PublicSubnetRouteTableAssociation"
	}
	return fmt.Sprintf("PublicSubnetRouteTableAssociation%02d", idx)
}

func VPCPeeringRouteName(idx int) string {
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	if idx < 1 {
		return "VPCPeeringRoute"
	}
	return fmt.Sprintf("VPCPeeringRoute%02d", idx)
}
