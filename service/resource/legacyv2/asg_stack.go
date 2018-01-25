package legacyv2

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/keyv2"
)

type asgStackInput struct {
	// Dependencies.
	clients awsutil.Clients

	// Settings.
	asgSize                int
	asgType                string
	availabilityZone       string
	bucket                 resources.ReusableResource
	cluster                v1alpha1.AWSConfig
	clusterID              string
	iamInstanceProfileName string
	imageID                string
	instanceType           string
	keyPairName            string
	loadBalancerName       string
	publicIP               bool
	subnetID               string
	tlsAssets              *legacy.CompactTLSAssets
	clusterKeys            *randomkeytpr.CompactRandomKeyAssets
	vpcID                  string
	workersSecurityGroupID string
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

func (s *Resource) processASGStack(input asgStackInput) (bool, error) {
	stack := awsresources.ASGStack{
		// Dependencies.
		Client: input.clients.CloudFormation,

		// Settings.
		Name: keyv2.AutoScalingGroupName(input.cluster, input.asgType),
	}

	stackExists, err := stack.CheckIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	var stackCreated bool

	if !stackExists {
		stackCreated, err = s.createASGStack(input)
		if err != nil {
			return stackCreated, microerror.Mask(err)
		}
	} else {
		stackCreated = true
		err = s.updateASGStack(input)
		if err != nil {
			return stackCreated, microerror.Mask(err)
		}
	}

	return stackCreated, nil
}

// createASGStack creates a CloudFormation stack for an Auto Scaling Group.
func (s *Resource) createASGStack(input asgStackInput) (bool, error) {
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

	var template string
	{
		var err error

		switch input.asgType {
		case prefixMaster:
			template, err = s.cloudConfig.NewMasterTemplate(input.cluster, *input.tlsAssets, *input.clusterKeys)
		case prefixWorker:
			template, err = s.cloudConfig.NewWorkerTemplate(input.cluster, *input.tlsAssets)
		default:
			return false, microerror.Maskf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.asgType))
		}

		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	cloudconfigConfig := SmallCloudconfigConfig{
		MachineType: input.asgType,
		Region:      input.cluster.Spec.AWS.Region,
		S3URI:       s.bucketName(input.cluster),
	}

	cloudconfigS3 := &awsresources.BucketObject{
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
		Name:      s.bucketObjectName(input.asgType),
		Data:      template,
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
		Name:                     keyv2.AutoScalingGroupName(input.cluster, input.asgType),
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

func (s *Resource) updateASGStack(input asgStackInput) error {
	var imageID string

	switch input.asgType {
	case prefixMaster:
		imageID = keyv2.MasterImageID(input.cluster)
	case prefixWorker:
		imageID = keyv2.WorkerImageID(input.cluster)
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
		Name:        keyv2.AutoScalingGroupName(input.cluster, input.asgType),
		TemplateURL: templateURL,
	}

	if err := stack.Update(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
