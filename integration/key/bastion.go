package key

import (
	"fmt"
)

func BastionIgnitionBucket(accountID string) string {
	return fmt.Sprintf("%s-bastion-ignition", accountID)
}

func BastionIgnitionObject(clusterID string) string {
	return fmt.Sprintf("%s.json", clusterID)
}

func BastionIgnitionURL(accountID string, clusterID string) string {
	return fmt.Sprintf("s3://%ss/%s.json", accountID, clusterID)
}
