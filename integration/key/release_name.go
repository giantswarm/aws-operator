// +build k8srequired

package key

import (
	"fmt"
)

func AWSConfigReleaseName(clusterID string) string {
	return fmt.Sprintf("e2esetup-awsconfig-%s", clusterID)
}

func AWSOperatorReleaseName() string {
	return "aws-operator"
}

func CertOperatorReleaseName() string {
	return "cert-operator"
}

func CertsReleaseName(clusterID string) string {
	return fmt.Sprintf("e2esetup-certs-%s", clusterID)
}

func CredentialdReleaseName() string {
	return "credentaild"
}

func NodeOperatorReleaseName() string {
	return "node-operator"
}

func VaultReleaseName() string {
	return "e2esetup-vault"
}
