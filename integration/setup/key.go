// +build k8srequired

package setup

import (
	"fmt"
)

func awsConfigReleaseName(clusterID string) string {
	return fmt.Sprintf("e2esetup-awsconfig-%s", clusterID)
}

func awsOperatorReleaseName() string {
	return "aws-operator"
}

func certOperatorReleaseName() string {
	return "cert-operator"
}

func certsReleaseName(clusterID string) string {
	return fmt.Sprintf("e2esetup-certs-%s", clusterID)
}

func nodeOperatorReleaseName() string {
	return "node-operator"
}

func vaultReleaseName() string {
	return "e2esetup-vault"
}
