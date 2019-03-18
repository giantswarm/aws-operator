package tccp

import "github.com/aws/aws-sdk-go/service/cloudformation"

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name string

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

	UpdateStackInput cloudformation.UpdateStackInput

	VersionBundleVersion string
}
