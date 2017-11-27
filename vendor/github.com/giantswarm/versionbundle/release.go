package versionbundle

import (
	"fmt"
	"sort"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

type ReleaseConfig struct {
	Bundles []Bundle
}

func DefaultReleaseConfig() ReleaseConfig {
	return ReleaseConfig{
		Bundles: nil,
	}
}

type Release struct {
	bundles    []Bundle
	changelogs []Changelog
	components []Component
	deprecated bool
	timestamp  string
	version    string
}

func NewRelease(config ReleaseConfig) (Release, error) {
	if len(config.Bundles) == 0 {
		return Release{}, microerror.Maskf(invalidConfigError, "config.Bundles must not be empty")
	}

	var err error

	var changelogs []Changelog
	{
		changelogs, err = aggregateReleaseChangelogs(config.Bundles)
		if err != nil {
			return Release{}, microerror.Maskf(invalidConfigError, err.Error())
		}
	}

	var components []Component
	{
		components, err = aggregateReleaseComponents(config.Bundles)
		if err != nil {
			return Release{}, microerror.Maskf(invalidConfigError, err.Error())
		}
	}

	var deprecated bool
	{
		deprecated, err = aggregateReleaseDeprecated(config.Bundles)
		if err != nil {
			return Release{}, microerror.Maskf(invalidConfigError, err.Error())
		}
	}

	var timestamp string
	{
		timestamp, err = aggregateReleaseTimestamp(config.Bundles)
		if err != nil {
			return Release{}, microerror.Maskf(invalidConfigError, err.Error())
		}
	}

	var version string
	{
		version, err = aggregateReleaseVersion(config.Bundles)
		if err != nil {
			return Release{}, microerror.Maskf(invalidConfigError, err.Error())
		}
	}

	r := Release{
		bundles:    config.Bundles,
		changelogs: changelogs,
		components: components,
		deprecated: deprecated,
		timestamp:  timestamp,
		version:    version,
	}

	return r, nil
}

func (r Release) Bundles() []Bundle {
	return CopyBundles(r.bundles)
}

func (r Release) Changelogs() []Changelog {
	return CopyChangelogs(r.changelogs)
}

func (r Release) Components() []Component {
	return CopyComponents(r.components)
}

func (r Release) Deprecated() bool {
	return r.deprecated
}

func (r Release) Timestamp() string {
	return r.timestamp
}

func (r Release) Version() string {
	return r.version
}

func aggregateReleaseChangelogs(bundles []Bundle) ([]Changelog, error) {
	var changelogs []Changelog

	for _, b := range bundles {
		changelogs = append(changelogs, b.Changelogs...)
	}

	return changelogs, nil
}

func aggregateReleaseComponents(bundles []Bundle) ([]Component, error) {
	var components []Component

	for _, b := range bundles {
		components = append(components, b.Components...)
	}

	return components, nil
}

func aggregateReleaseDeprecated(bundles []Bundle) (bool, error) {
	for _, b := range bundles {
		if b.Deprecated == true {
			return true, nil
		}
	}

	return false, nil
}

func aggregateReleaseTimestamp(bundles []Bundle) (string, error) {
	var t time.Time

	for _, b := range bundles {
		if b.Time.After(t) {
			t = b.Time
		}
	}

	return t.Format("2006-01-02T15:04:05.000000Z"), nil
}

func aggregateReleaseVersion(bundles []Bundle) (string, error) {
	var major int64
	var minor int64
	var patch int64

	for _, b := range bundles {
		v, err := semver.NewVersion(b.Version)
		if err != nil {
			return "", microerror.Mask(err)
		}

		major += v.Major
		minor += v.Minor
		patch += v.Patch
	}

	version := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	return version, nil
}

func GetNewestRelease(releases []Release) (Release, error) {
	if len(releases) == 0 {
		return Release{}, microerror.Maskf(executionFailedError, "releases must not be empty")
	}

	s := SortReleasesByVersion(releases)
	sort.Sort(s)

	return s[len(s)-1], nil
}
