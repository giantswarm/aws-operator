package versionbundle

import "time"

type SortReleasesByVersion []Release

func (r SortReleasesByVersion) Len() int           { return len(r) }
func (r SortReleasesByVersion) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r SortReleasesByVersion) Less(i, j int) bool { return r[i].Version() < r[j].Version() }

type SortReleasesByTimestamp []Release

func (r SortReleasesByTimestamp) Len() int      { return len(r) }
func (r SortReleasesByTimestamp) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r SortReleasesByTimestamp) Less(i, j int) bool {
	iTime, err := time.Parse(releaseTimestampFormat, r[i].Timestamp())
	if err != nil {
		panic(err)
	}

	jTime, err := time.Parse(releaseTimestampFormat, r[j].Timestamp())
	if err != nil {
		panic(err)
	}

	return iTime.UnixNano() < jTime.UnixNano()
}
