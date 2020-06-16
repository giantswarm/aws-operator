package unittest

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultRelease() releasev1alpha1.Release {
	cr := releasev1alpha1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name: "v100.0.0",
		},
		Spec: releasev1alpha1.ReleaseSpec{
			Components: []releasev1alpha1.ReleaseSpecComponent{
				{
					Name:    "calico",
					Version: "3.10.1",
				},
				{
					Name:    "containerlinux",
					Version: "2345.3.1",
				},
				{
					Name:    "etcd",
					Version: "3.4.9",
				},
				{
					Name:    "kubernetes",
					Version: "1.16.9",
				},
			},
		},
	}

	return cr
}
