package versionbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const indexReleaseTimestampFormat = "2006-01-02T15:04:05.00Z"

type IndexRelease struct {
	Active      bool        `yaml:"active"`
	Authorities []Authority `yaml:"authorities"`
	Date        time.Time   `yaml:"date"`
	Version     string      `yaml:"version"`
}

// CompileReleases takes indexReleases and collected version bundles and
// compiles canonicalized Releases from them.
func CompileReleases(logger micrologger.Logger, indexReleases []IndexRelease, bundles []Bundle) ([]Release, error) {
	releases, err := buildReleases(logger, indexReleases, bundles)
	if err != nil {
		return nil, err
	}

	releases = deduplicateReleaseChangelog(releases)

	return releases, nil
}

func buildReleases(logger micrologger.Logger, indexReleases []IndexRelease, bundles []Bundle) ([]Release, error) {
	bundleCache := make(map[string]Bundle)

	// Create cache of bundles for quick lookup
	for _, b := range bundles {
		bundleCache[b.ID()] = b
	}

	var releases []Release

	for _, ir := range indexReleases {
		bundles, err := groupBundlesForIndexRelease(ir, bundleCache)
		if IsBundleNotFound(err) {
			logger.Log("level", "warning", "message", fmt.Sprintf("failed grouping version bundles for release %s", ir.Version), "stack", fmt.Sprintf("%#v", err))
			continue
		}

		if err != nil {
			return nil, err
		}

		rc := ReleaseConfig{
			Active:  ir.Active,
			Bundles: bundles,
			Date:    ir.Date,
			Version: ir.Version,
		}

		release, err := NewRelease(rc)
		if err != nil {
			logger.Log("level", "warning", "message", fmt.Sprintf("failed building new release from %s", ir.Version), "stack", fmt.Sprintf("%#v", err))
			continue
		}

		releases = append(releases, release)
	}

	return releases, nil
}

func groupBundlesForIndexRelease(ir IndexRelease, bundles map[string]Bundle) ([]Bundle, error) {
	var groupedBundles []Bundle
	for _, a := range ir.Authorities {
		b, found := bundles[a.BundleID()]
		if !found {
			return nil, microerror.Maskf(bundleNotFoundError, "IndexRelease v%s contains Authority with bundle ID %s that cannot be found from collected version bundles.", ir.Version, a.BundleID())
		}
		groupedBundles = append(groupedBundles, b)
	}

	return groupedBundles, nil
}

// deduplicateReleaseChangelog removes duplicate changelog entries in
// consecutive release entries. Core concept of algorithm here is to first sort
// releases by their release version and then iterate them and compare current
// release to previous one that fulfills following requirements: smaller
// version number and earlier timestamp. Comparison of earlier timestamp is
// crucial here in order to calculate changelog correctly when newer patch
// releases have been introduced with lower version number
// (e.g. [1.0.0, 2.0.0] -> [1.0.0, 1.0.1, 2.0.0, 2.0.1]).
func deduplicateReleaseChangelog(releases []Release) []Release {
	if len(releases) < 2 {
		return releases
	}

	sort.Sort(SortReleasesByVersion(releases))

	var filteredReleases []Release

	for i := 0; i < len(releases); i++ {
		r := releases[i]

		// Deepcopy changelogs as they are modified later and this instance of
		// r ends up to filteredReleases.
		r.changelogs = make([]Changelog, len(releases[i].changelogs))
		copy(r.changelogs, releases[i].changelogs)

		// Find previous release and map changelogs for quick lookup.
		prevRelease := findPreviousRelease(r, releases[:i])
		prevChangelogs := mapReleaseChangelogs(prevRelease)

		// Process changelogs of current release removing ones present in
		// previous release.
		for _, clog := range r.Changelogs() {
			_, exists := prevChangelogs[clog.String()]
			if exists {
				// r.Changelogs() returns a copy of changelogs so removal won't
				// mess iteration in this case.
				r.removeChangelogEntry(clog)
			}

		}

		filteredReleases = append(filteredReleases, r)
	}

	return filteredReleases
}

// findPreviousRelease finds release that is older than argument r0. This
// function expects that releases is sorted by version as it is iterated
// backwards. If no previous release is found, empty Release is returned.
func findPreviousRelease(r0 Release, releases []Release) Release {
	for i := len(releases) - 1; i >= 0; i-- {
		if releases[i].timestamp.Before(r0.timestamp) {
			return releases[i]
		}
	}

	return Release{}
}

// mapReleaseChangelogs converts Release's Changelog slice to
// map[string]struct{} for quick lookup.
func mapReleaseChangelogs(r Release) map[string]struct{} {
	changelogs := make(map[string]struct{})
	for _, clog := range r.Changelogs() {
		changelogs[clog.String()] = struct{}{}
	}

	return changelogs
}

// ValidateIndexReleases ensures semantic rules for collection of indexReleases
// so that when used together, they form consistent and integral release index.
func ValidateIndexReleases(indexReleases []IndexRelease) error {
	if len(indexReleases) == 0 {
		return nil
	}

	var err error

	err = validateReleaseAuthorities(indexReleases)
	if err != nil {
		return microerror.Mask(err)
	}
	err = validateReleaseDates(indexReleases)
	if err != nil {
		return microerror.Mask(err)
	}
	err = validateUniqueReleases(indexReleases)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func validateReleaseAuthorities(indexReleases []IndexRelease) error {
	for _, release := range indexReleases {
		if len(release.Authorities) == 0 {
			return microerror.Maskf(invalidReleaseError, "release %s has no authorities", release.Version)
		}

		for _, authority := range release.Authorities {
			if authority.Name == "" {
				return microerror.Maskf(invalidReleaseError, "release %s contains authority without Name", release.Version)
			}

			if authority.Endpoint == nil {
				return microerror.Maskf(invalidReleaseError, "release %s authority %s doesn't have defined endpoint", release.Version, authority.Name)
			}

			if authority.Version == "" {
				return microerror.Maskf(invalidReleaseError, "release %s authority %s doesn't have defined version", release.Version, authority.Name)
			}
		}
	}
	return nil
}

func validateReleaseDates(indexReleases []IndexRelease) error {
	releaseDates := make(map[time.Time]string)
	for _, release := range indexReleases {
		if release.Date.IsZero() {
			return microerror.Maskf(invalidReleaseError, "release %s has empty release date", release.Version)
		}

		releaseDates[release.Date] = release.Version
	}

	return nil
}

func validateUniqueReleases(indexReleases []IndexRelease) error {
	releaseChecksums := make(map[string]string)
	releaseVersions := make(map[string]string)

	sha256Hash := sha256.New()

	for _, release := range indexReleases {
		// Verify release version number
		otherVer, exists := releaseVersions[release.Version]
		if exists {
			return microerror.Maskf(invalidReleaseError, "duplicate release versions %s and %s", otherVer, release.Version)
		}

		releaseVersions[release.Version] = release.Version

		// Verify release version contents
		authorities := make([]string, 0, len(release.Authorities))
		for _, a := range release.Authorities {
			authorities = append(authorities, a.BundleID())
		}

		sort.Strings(authorities)

		sha256Hash.Reset()
		sha256Hash.Write([]byte(strings.Join(authorities, ",")))

		hexHash := hex.EncodeToString(sha256Hash.Sum(nil))
		otherVer, exists = releaseChecksums[hexHash]
		if exists {
			return microerror.Maskf(invalidReleaseError, "duplicate release contents for versions %s and %s", otherVer, release.Version)
		}
		releaseChecksums[hexHash] = release.Version
	}

	return nil
}
