package changedetection

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
)

//components which matter to trigger update
// kubernetes

func releaseComponentsEqual(currentRelease releasev1alpha1.Release, targetRelease releasev1alpha1.Release) bool {
	// TODO
	return false
}
