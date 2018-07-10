package adapter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	prefixMaster  = "master"
	prefixWorker  = "worker"
	prefixIngress = "ingress"

	// asgMaxBatchSizeRatio is the % of instances to be updated during a
	// rolling update.
	asgMaxBatchSizeRatio = 0.3
	// asgMinInstancesRatio is the % of instances to keep in service during a
	// rolling update.
	asgMinInstancesRatio = 0.7
	// defaultEBSVolumeMountPoint is the path for mounting the EBS volume.
	defaultEBSVolumeMountPoint = "/dev/xvdh"
	// defaultEBSVolumeSize is expressed in GB.
	defaultEBSVolumeSize = 100
	// defaultEBSVolumeType is the EBS volume type.
	defaultEBSVolumeType = "gp2"
	// rollingUpdatePauseTime is how long to pause ASG operations after creating
	// new instances. This allows time for new nodes to join the cluster.
	rollingUpdatePauseTime = "PT15M"

	// Subnet keys
	subnetDescription = "description"
	subnetGroupName   = "group-name"

	// accountIDIndex represents the index in which we can find the account ID in the user ARN
	// (splitting by colon)
	accountIDIndex  = 4
	accountIDLength = 12

	// The number of seconds AWS will wait, before issuing a health check on
	// instances in an Auto Scaling Group.
	gracePeriodSeconds = 10

	tagKeyName = "Name"

	suffixPublic  = "public"
	suffixPrivate = "private"

	externalELBScheme = "internet-facing"
	internalELBScheme = "internal"

	httpPort  = 80
	httpsPort = 443

	// RootDirElement marks the directory that should be taken as root when evaluating
	// template's relative paths.
	RootDirElement = "aws-operator"
)

// APIWhitelist defines guest cluster k8s api whitelisting.
type APIWhitelist struct {
	Enabled    bool
	SubnetList string
}

type Clients struct {
	CloudFormation CFClient
	EC2            EC2Client
	IAM            IAMClient
	KMS            KMSClient
	ELB            ELBClient
	STS            STSClient
}

// TODO we copy this because of a circular import issue with the cloudformation
// resource. The way how the resource works with the adapter and how infromation
// is passed has to be reworked at some point. Just hacking this now to keep
// going and to keep the changes as minimal as possible.
type StackState struct {
	Name string

	MasterImageID              string
	MasterInstanceType         string
	MasterInstanceResourceName string
	MasterCloudConfigVersion   string
	MasterInstanceMonitoring   bool

	WorkerCount              string
	WorkerImageID            string
	WorkerInstanceMonitoring bool
	WorkerInstanceType       string
	WorkerCloudConfigVersion string

	VersionBundleVersion string
}

// CFClient describes the methods required to be implemented by a CloudFormation
// AWS client.
type CFClient interface {
	CreateStack(*awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error)
	DeleteStack(*awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error)
	DescribeStacks(*awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error)
	UpdateStack(*awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error)
	WaitUntilStackCreateComplete(*awscloudformation.DescribeStacksInput) error
	WaitUntilStackCreateCompleteWithContext(ctx aws.Context, input *awscloudformation.DescribeStacksInput, opts ...request.WaiterOption) error
}

// EC2Client describes the methods required to be implemented by a EC2 AWS client.
type EC2Client interface {
	DescribeAddresses(*ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error)
	DescribeSecurityGroups(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeSubnets(*ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)
	DescribeRouteTables(*ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error)
	DescribeVpcs(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error)
	DescribeVpcPeeringConnections(*ec2.DescribeVpcPeeringConnectionsInput) (*ec2.DescribeVpcPeeringConnectionsOutput, error)
}

// IAMClient describes the methods required to be implemented by a IAM AWS client.
type IAMClient interface {
	GetUser(*iam.GetUserInput) (*iam.GetUserOutput, error)
	GetRole(*iam.GetRoleInput) (*iam.GetRoleOutput, error)
}

// KMSClient describes the methods required to be implemented by a KMS AWS client.
type KMSClient interface {
	DescribeKey(*kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error)
}

// SmallCloudconfigConfig represents the data structure required for executing the
// small cloudconfig template.
type SmallCloudconfigConfig struct {
	MachineType        string
	Region             string
	S3URI              string
	CloudConfigVersion string
}

// ELBClient describes the methods required to be implemented by a ELB AWS client.
type ELBClient interface {
	DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
}

// STSClient describes the methods required to be implemented by a STS AWS client.
type STSClient interface {
	GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}
