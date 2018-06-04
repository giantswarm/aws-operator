package versionbundle

import (
	"github.com/coreos/go-semver/semver"
)

type SortIndexReleasesByVersion []IndexRelease

func (r SortIndexReleasesByVersion) Len() int      { return len(r) }
func (r SortIndexReleasesByVersion) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r SortIndexReleasesByVersion) Less(i, j int) bool {
	verA := semver.New(r[i].Version)
	verB := semver.New(r[j].Version)
	return verA.LessThan(*verB)
}
