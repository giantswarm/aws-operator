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
