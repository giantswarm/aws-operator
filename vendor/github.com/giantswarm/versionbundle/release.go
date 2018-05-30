package versionbundle

import (
	"sort"
	"time"

	"github.com/giantswarm/microerror"
)

const releaseTimestampFormat = "2006-01-02T15:04:05.000000Z"

type ReleaseConfig struct {
	Active  bool
	Bundles []Bundle
	Date    time.Time
	Version string
}

type Release struct {
	bundles    []Bundle
	changelogs []Changelog
	components []Component
	timestamp  time.Time
	version    string
	active     bool
}

func NewRelease(config ReleaseConfig) (Release, error) {
	if len(config.Bundles) == 0 {
		return Release{}, microerror.Maskf(invalidConfigError, "%T.Bundles must not be empty", config)
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

	r := Release{
		active:     config.Active,
		bundles:    config.Bundles,
		changelogs: changelogs,
		components: components,
		timestamp:  config.Date,
		version:    config.Version,
	}

	return r, nil
}

func (r Release) Active() bool {
	return r.active
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

func (r Release) Timestamp() string {
	if r.timestamp.IsZero() {
		// This maintains existing behavior.
		return ""
	}

	return r.timestamp.Format(releaseTimestampFormat)
}

func (r Release) Version() string {
	return r.version
}

func (r *Release) removeChangelogEntry(clog Changelog) {
	for i := 0; i < len(r.changelogs); i++ {
		if clog == r.changelogs[i] {
			r.changelogs = append(r.changelogs[:i], r.changelogs[i+1:]...)
			break
		}
	}

	return
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

func GetNewestRelease(releases []Release) (Release, error) {
	if len(releases) == 0 {
		return Release{}, microerror.Maskf(executionFailedError, "releases must not be empty")
	}

	s := SortReleasesByVersion(releases)
	sort.Sort(s)

	return s[len(s)-1], nil
}
