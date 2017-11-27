package versionbundle

type SortReleasesByVersion []Release

func (r SortReleasesByVersion) Len() int           { return len(r) }
func (r SortReleasesByVersion) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r SortReleasesByVersion) Less(i, j int) bool { return r[i].Version() < r[j].Version() }
