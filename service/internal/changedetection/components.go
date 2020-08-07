package changedetection

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
)

//components which matter to trigger update
// Kubernetes, aws-cni, ...

var components = []string{"kubernetes", "etcd", "aws-cni", "calico", "containerlinux"}

func releaseComponentsEqual(currentRelease releasev1alpha1.Release, targetRelease releasev1alpha1.Release) bool {
	if len(currentRelease.Spec.Components) == 0 {
		return false
	}
	for _, current := range currentRelease.Spec.Components {
		if findComponent(current.Name) {
			for _, target := range targetRelease.Spec.Components {
				if current.Name == target.Name {
					if current.Version != target.Version {
						return false
					}
				}
			}

		}
	}
	return true
}

func findComponent(val string) bool {
	for _, c := range components {
		if val == c {
			return true
		}
	}
	return false
}