package versionbundle

import "sort"

// CanonicalizeReleases canonicalizes representation of multiple releases. It
// makes sure there are no duplicate versions.
func CanonicalizeReleases(releases []Release) []Release {
	releases = deduplicateReleases(releases)
	// TODO: Add changelog deduplication here.

	return releases
}

func deduplicateReleases(releases []Release) []Release {
	if len(releases) < 2 {
		return releases
	}

	sort.Sort(SortReleasesByTimestamp(releases))
	sort.Stable(SortReleasesByVersion(releases))

	for i := 0; i < len(releases)-1; i++ {
		if releases[i].Version() == releases[i+1].Version() {
			// NOTE: This will remove newer version of two duplicates to
			// maintain active version available.
			if !releases[i].Active() {
				releases = append(releases[:i], releases[i+1:]...)
				i--
			} else {
				releases = append(releases[:i+1], releases[i+2:]...)
				i--
			}
		}
	}

	return releases
}
