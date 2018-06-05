package versionbundle

type SortComponentsByName []Component

func (c SortComponentsByName) Len() int           { return len(c) }
func (c SortComponentsByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c SortComponentsByName) Less(i, j int) bool { return c[i].Name < c[j].Name }
