package unittest

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultNetworkPool(cidr string) infrastructurev1alpha3.NetworkPool {
	cr := infrastructurev1alpha3.NetworkPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha3.NetworkPoolSpec{
			CIDRBlock: cidr,
		},
	}

	return cr
}
