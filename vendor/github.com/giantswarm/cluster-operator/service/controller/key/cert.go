package key

import (
	"fmt"
)

const (
	LocalhostIP = "127.0.0.1"
)

// CertDefaultAltNames returns default alt names for Kubernetes API certs.
func CertDefaultAltNames(clusterDomain string) []string {
	return []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		fmt.Sprintf("kubernetes.default.svc.%s", clusterDomain),
	}
}

// CertConfigName constructs a name for CertConfig CRs using the clusterI D and
// the cert name.
func CertConfigName(getter LabelsGetter, name string) string {
	return fmt.Sprintf("%s-%s", ClusterID(getter), name)
}
