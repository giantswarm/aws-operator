package template

type ParamsMainEtcdVolume struct {
	List []ParamsMainEtcdVolumeItem
}

type ParamsMainEtcdVolumeItem struct {
	AvailabilityZone string
	Iops             int
	Name             string
	Resource         string
	SnapshotID       string
	Throughput       int
}
