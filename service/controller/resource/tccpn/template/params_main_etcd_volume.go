package template

type ParamsMainEtcdVolume struct {
	List []ParamsMainEtcdVolumeItem
}

type ParamsMainEtcdVolumeItem struct {
	AvailabilityZone string
	Name             string
	Resource         string
	SnapshotID       string
}
