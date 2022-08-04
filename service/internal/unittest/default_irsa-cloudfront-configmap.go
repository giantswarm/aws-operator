package unittest

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultIRSACloudfrontConfigMap() v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-irsa-cloudfront", DefaultClusterID),
			Namespace: metav1.NamespaceDefault,
		},
		Data: map[string]string{
			"domain": "122424fd.cloudfront.net",
		},
	}
}
