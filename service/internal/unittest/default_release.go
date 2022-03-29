package unittest

import (
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"
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
					Name:    "containerlinux",
					Version: "2345.3.1",
				},
			},
		},
	}

	return cr
}
