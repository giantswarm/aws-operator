package unittest

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultNetworkPool(cidr string) infrastructurev1alpha2.NetworkPool {
	cr := infrastructurev1alpha2.NetworkPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.NetworkPoolSpec{
			CIDRBlock: cidr,
		},
	}

	return cr
}
