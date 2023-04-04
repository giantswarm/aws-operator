package changedetection

import (
	"fmt"

	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
)

var components = []string{"kubernetes", "etcd", "calico", "containerlinux"}

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

func componentsDiff(currentRelease releasev1alpha1.Release, targetRelease releasev1alpha1.Release) []string {
	var diff []string

	for _, current := range currentRelease.Spec.Components {
		if findComponent(current.Name) {
			for _, target := range targetRelease.Spec.Components {
				if current.Name == target.Name {
					if current.Version != target.Version {
						diff = append(diff, fmt.Sprintf("%s version changed from %s to %s", current.Name, current.Version, target.Version))
					}
				}
			}
		}
	}

	return diff
}

func findComponent(val string) bool {
	for _, c := range components {
		if val == c {
			return true
		}
	}
	return false
}
