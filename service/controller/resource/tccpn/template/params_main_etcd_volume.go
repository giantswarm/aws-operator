package template

type ParamsMainEtcdVolume struct {
	Volumes []ParamsMainEtcdVolumeEtcdVolumeSpec
}

type ParamsMainEtcdVolumeEtcdVolumeSpec struct {
	AvailabilityZone string
	Name             string
	SnapshotID       string
	ResourceName     string
}
