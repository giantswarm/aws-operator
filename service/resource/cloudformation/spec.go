package cloudformation

import (
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the stack.
	defaultCreationTimeout = 10

	workersParameterKey        = "workers"
	imageIDParameterKey        = "imageID"
	clusterVersionParameterKey = "clusterVersion"

	cloudFormationTemplatesDirectory = "service/templates/cloudformation"

	smallCloudConfigTemplate = "service/templates/cloudconfig/small_cloudconfig.yaml"

	prefixWorker = "worker"
	// asgMaxBatchSizeRatio is the % of instances to be updated during a
	// rolling update.
	asgMaxBatchSizeRatio = 0.3
	// asgMinInstancesRatio is the % of instances to keep in service during a
	// rolling update.
	asgMinInstancesRatio = 0.7
	// defaultEBSVolumeMountPoint is the path for mounting the EBS volume.
	defaultEBSVolumeMountPoint = "/dev/xvdh"
	// defaultEBSVolumeSize is expressed in GB.
	defaultEBSVolumeSize = 50
	// defaultEBSVolumeType is the EBS volume type.
	defaultEBSVolumeType = "gp2"
	// rollingUpdatePauseTime is how long to pause ASG operations after creating
	// new instances. This allows time for new nodes to join the cluster.
	rollingUpdatePauseTime = "PT5M"

	// Subnet keys
	subnetDescription = "description"
	subnetGroupName   = "group-name"

	// accountIDIndex represents the index in which we can find the account ID in the user ARN
	// (splitting by colon)
	accountIDIndex  = 4
	accountIDLength = 12
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name           string
	ImageID        string
	Workers        string
	ClusterVersion string
}

// EC2Client describes the methods required to be implemented by a EC2 AWS client.
type EC2Client interface {
	DescribeSecurityGroups(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)
}

// CFClient describes the methods required to be implemented by a CloudFormation AWS client.
type CFClient interface {
	CreateStack(*awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error)
	DeleteStack(*awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error)
	DescribeStacks(*awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error)
	UpdateStack(*awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error)
}

// IAMClient describes the methods required to be implemented by a IAM AWS client.
type IAMClient interface {
	GetUser(*iam.GetUserInput) (*iam.GetUserOutput, error)
}

// SmallCloudconfigConfig represents the data structure required for executing the
// small cloudconfig template.
type SmallCloudconfigConfig struct {
	MachineType string
	Region      string
	S3URI       string
}
