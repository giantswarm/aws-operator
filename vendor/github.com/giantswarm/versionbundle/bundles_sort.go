package versionbundle

type SortBundlesByName []Bundle

func (b SortBundlesByName) Len() int           { return len(b) }
func (b SortBundlesByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByName) Less(i, j int) bool { return b[i].Name < b[j].Name }

type SortBundlesByVersion []Bundle

func (b SortBundlesByVersion) Len() int           { return len(b) }
func (b SortBundlesByVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }

type SortBundlesByTime []Bundle

func (b SortBundlesByTime) Len() int           { return len(b) }
func (b SortBundlesByTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByTime) Less(i, j int) bool { return b[i].Time.UnixNano() < b[j].Time.UnixNano() }
