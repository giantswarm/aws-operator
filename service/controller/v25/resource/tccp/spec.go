package tccp

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name     string
	Template string

	// NOTE everything below is deprecated. We try to cleanup the state being
	// dispatched between resource operations.

	DockerVolumeResourceName   string
	MasterImageID              string
	MasterInstanceType         string
	MasterInstanceResourceName string
	MasterCloudConfigVersion   string
	MasterInstanceMonitoring   bool

	ShouldScale  bool
	ShouldUpdate bool

	WorkerCloudConfigVersion string
	WorkerDockerVolumeSizeGB string
	WorkerImageID            string
	WorkerInstanceMonitoring bool
	WorkerInstanceType       string

	VersionBundleVersion string
}
