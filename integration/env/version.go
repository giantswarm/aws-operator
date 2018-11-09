package env

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

func compareVersions(v1 string, v2 string) (int, error) {
	s1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, microerror.Mask(err)
	}
	s2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return s1.Compare(*s2), nil
}

func isIPAMEnabled() bool {
	n, err := compareVersions(VersionBundleVersion(), "4.4.0")
	if err != nil {
		panic(fmt.Sprintf("%#v", microerror.Mask(err)))
	}

	if n == 0 || n == 1 {
		// TODO the version bundle version of the AWSConfig CR is equal to or bigger
		// than 4.4.0. This version implies the resource package v19. Thus we do not
		// want to set the deprecated network configurations in the CR. We track
		// removing ths check in the roadmap issue.
		//
		//     https://github.com/giantswarm/giantswarm/pull/2202
		//
		return true
	}

	return false
}
