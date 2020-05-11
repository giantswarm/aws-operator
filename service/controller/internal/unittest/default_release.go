package unittest

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
)

func DefaultRelease() releasev1alpha1.Release {
	cr := releasev1alpha1.Release{
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
