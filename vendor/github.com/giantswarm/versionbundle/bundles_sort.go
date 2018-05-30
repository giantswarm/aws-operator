package versionbundle

import "github.com/coreos/go-semver/semver"

type SortBundlesByName []Bundle

func (b SortBundlesByName) Len() int           { return len(b) }
func (b SortBundlesByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByName) Less(i, j int) bool { return b[i].Name < b[j].Name }

type SortBundlesByVersion []Bundle

func (b SortBundlesByVersion) Len() int      { return len(b) }
func (b SortBundlesByVersion) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByVersion) Less(i, j int) bool {
	verA := semver.New(b[i].Version)
	verB := semver.New(b[j].Version)
	return verA.LessThan(*verB)
}
