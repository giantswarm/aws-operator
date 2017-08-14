package create

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/key"
)

type asgStackInput struct {

	// Settings.
	asgSize                int
	asgType                string
	availabilityZone       string
	bucket                 resources.ReusableResource
	cluster                awstpr.CustomObject
	clusterID              string
	iamInstanceProfileName string
	imageID                string
	instanceType           string
	keyPairName            string
	loadBalancerName       string
	publicIP               bool
	subnetID               string
	tlsAssets              *certificatetpr.CompactTLSAssets
	vpcID                  string
	workersSecurityGroupID string

	// Dependencies.
	clients awsutil.Clients
}

type blockDeviceMapping struct {
	DeviceName          string
	DeleteOnTermination bool
	VolumeSize          int64
	VolumeType          string
}

type asgTemplateConfig struct {
	ASGType             string
	BlockDeviceMappings []blockDeviceMapping
}

const (
	// asgCloudFormationGoTemplate is the Go template that generates the Cloud
	// Formation template.
	asgCloudFormationGoTemplate = "resources/templates/cloudformation/auto_scaling_group.yaml"
	// asgCloudFormationTemplateS3Path is the path to the Cloud Formation
	// template stored in the S3 bucket.
	asgCloudFormationTemplateS3Path = "templates/%s.yaml"
	// defaultEBSVolumeMountPoint is the path for mounting the EBS volume.
	defaultEBSVolumeMountPoint = "/dev/xvdh"
	// defaultEBSVolumeSize is expressed in GB.
	defaultEBSVolumeSize = 50
	// defaultEBSVolumeType is the EBS volume type.
	defaultEBSVolumeType = "gp2"
)

// createASGStack creates a CloudFormation stack for an Auto Scaling Group.
func (s *Service) createASGStack(input asgStackInput) (bool, error) {
	var (
		extension cloudconfig.Extension
	)

	// Generate the Cloud Formation template using a Go template.
	cfTemplate, err := ioutil.ReadFile(asgCloudFormationGoTemplate)
	if err != nil {
		return false, microerror.Mask(err)
	}

	goTemplate, err := template.New("asg").Parse(string(cfTemplate))
	if err != nil {
		return false, microerror.Mask(err)
	}

	parsedTemplate := new(bytes.Buffer)
	tc := asgTemplateConfig{
		ASGType: input.asgType,
		BlockDeviceMappings: []blockDeviceMapping{
			{
				DeviceName:          defaultEBSVolumeMountPoint,
				DeleteOnTermination: true,
				VolumeSize:          defaultEBSVolumeSize,
				VolumeType:          defaultEBSVolumeType,
			},
		},
	}

	if err := goTemplate.Execute(parsedTemplate, tc); err != nil {
		return false, microerror.Mask(err)
	}

	// Upload the Cloud Formation template to the S3 bucket.
	templateRelativePath := fmt.Sprintf(asgCloudFormationTemplateS3Path, input.asgType)
	templateURL := s.bucketObjectURL(input.cluster, templateRelativePath)

	templateS3 := &awsresources.BucketObject{
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
		Name:      templateRelativePath,
		Data:      parsedTemplate.String(),
		Bucket:    input.bucket.(*awsresources.Bucket),
	}
	if err := templateS3.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	switch input.asgType {
	case prefixMaster:
		extension = NewMasterCloudConfigExtension(input.cluster.Spec, input.tlsAssets)
	case prefixWorker:
		extension = NewWorkerCloudConfigExtension(input.cluster.Spec, input.tlsAssets)
	default:
		return false, microerror.Maskf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.asgType))
	}

	cloudConfigParams := cloudconfig.Params{
		Cluster:   input.cluster.Spec.Cluster,
		Extension: extension,
	}

	// We now upload the instance cloudconfig to S3 and create a "small
	// cloudconfig" that just fetches the previously uploaded "final
	// cloudconfig" and executes coreos-cloudinit with it as argument.
	// We do this to circumvent the 16KB limit on user-data for EC2 instances.
	cloudConfig, err := s.cloudConfig(prefixWorker, cloudConfigParams, input.cluster.Spec, input.tlsAssets)
	if err != nil {
		return false, microerror.Mask(err)
	}

	cloudconfigConfig := SmallCloudconfigConfig{
		MachineType: input.asgType,
		Region:      input.cluster.Spec.AWS.Region,
		S3URI:       s.bucketName(input.cluster),
	}

	cloudconfigS3 := &awsresources.BucketObject{
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
		Name:      s.bucketObjectName(input.asgType),
		Data:      cloudConfig,
		Bucket:    input.bucket.(*awsresources.Bucket),
	}
	if err := cloudconfigS3.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// Create CloudFormation stack for the ASG.
	stack := awsresources.ASGStack{
		// Dependencies.
		Client: input.clients.CloudFormation,

		// Settings.
		ASGMaxSize:               input.asgSize,
		ASGMinSize:               input.asgSize,
		ASGType:                  input.asgType,
		AssociatePublicIPAddress: input.publicIP,
		AvailabilityZone:         input.availabilityZone,
		ClusterID:                input.clusterID,
		HealthCheckGracePeriod:   gracePeriodSeconds,
		IAMInstanceProfileName:   input.iamInstanceProfileName,
		ImageID:                  input.imageID,
		LoadBalancerName:         input.loadBalancerName,
		InstanceType:             input.instanceType,
		KeyName:                  input.keyPairName,
		Name:                     key.AutoScalingGroupName(input.cluster, input.asgType),
		SecurityGroupID:          input.workersSecurityGroupID,
		SmallCloudConfig:         smallCloudconfig,
		SubnetID:                 input.subnetID,
		TemplateURL:              templateURL,
		VPCID:                    input.vpcID,
	}

	err = stack.CreateOrFail()
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (s *Service) updateASGStack(input asgStackInput) error {
	var imageID string

	switch input.asgType {
	case prefixMaster:
		imageID = key.MasterImageID(input.cluster)
	case prefixWorker:
		imageID = key.WorkerImageID(input.cluster)
	default:
		return microerror.Maskf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.asgType))
	}

	templateRelativePath := fmt.Sprintf(asgCloudFormationTemplateS3Path, input.asgType)
	templateURL := s.bucketObjectURL(input.cluster, templateRelativePath)

	// Update CloudFormation stack for the ASG.
	stack := awsresources.ASGStack{
		// Dependencies.
		Client: input.clients.CloudFormation,

		// Settings.
		ASGMaxSize:  input.asgSize,
		ASGMinSize:  input.asgSize,
		ImageID:     imageID,
		Name:        key.AutoScalingGroupName(input.cluster, input.asgType),
		TemplateURL: templateURL,
	}

	if err := stack.Update(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
