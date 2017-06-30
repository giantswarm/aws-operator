package create

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
)

// TODO rename to ASGStackInput
type asgStackInput struct {
	// Settings.
	bucket                    resources.ReusableResource
	cluster                   awstpr.CustomObject
	ebsStorage                bool
	iamInstanceProfileName    string
	elbListeners              []awsresources.PortPair
	keyPairName               string
	prefix                    string
	workersSecurityGroupRules []awsresources.SecurityGroupRule
	ingressSecurityGroupRules []awsresources.SecurityGroupRule
	subnetID                  string
	tlsAssets                 *certificatetpr.CompactTLSAssets
	vpcID                     string

	// Dependencies.
	clients awsutil.Clients
	logger  micrologger.Logger
}

type asgTemplateConfig struct {
	WorkersSecurityGroupRules []awsresources.SecurityGroupRule
	IngressSecurityGroupRules []awsresources.SecurityGroupRule
	Listeners                 []awsresources.PortPair
}

// createASGStack creates a CloudFormation stack for an Auto Scaling Group.
func (s *Service) createASGStack(input asgStackInput) error {
	var (
		extension    cloudconfig.Extension
		imageID      string
		instanceType string
		publicIP     bool
	)

	switch input.prefix {
	case prefixMaster:
		extension = NewMasterCloudConfigExtension(input.cluster.Spec, input.tlsAssets)

		// TODO Check only a single master node is provided.
		imageID = input.cluster.Spec.AWS.Masters[0].ImageID
		instanceType = input.cluster.Spec.AWS.Masters[0].InstanceType
	case prefixWorker:
		extension = NewWorkerCloudConfigExtension(input.cluster.Spec, input.tlsAssets)

		imageID = input.cluster.Spec.AWS.Workers[0].ImageID
		instanceType = input.cluster.Spec.AWS.Workers[0].InstanceType
		publicIP = true
	default:
		return microerror.MaskAnyf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.prefix))
	}

	// Upload the CF template to an S3 bucket.
	cfTemplate, err := ioutil.ReadFile("resources/cloudformation/auto_scaling_group.yaml")
	if err != nil {
		return microerror.MaskAny(err)
	}

	goTemplate, err := template.New("asg").Parse(string(cfTemplate))
	if err != nil {
		return microerror.MaskAny(err)
	}

	parsedTemplate := new(bytes.Buffer)
	tc := asgTemplateConfig{
		WorkersSecurityGroupRules: input.workersSecurityGroupRules,
		IngressSecurityGroupRules: input.ingressSecurityGroupRules,
		Listeners:                 input.elbListeners,
	}

	if err := goTemplate.Execute(parsedTemplate, tc); err != nil {
		return microerror.MaskAny(err)
	}

	templateRelativePath := fmt.Sprintf("templates/%s.yaml", input.prefix)

	templateURL := s.bucketObjectURL(input.cluster, templateRelativePath)

	templateS3 := &awsresources.BucketObject{
		Name:   templateRelativePath,
		Data:   parsedTemplate.String(),
		Bucket: input.bucket.(*awsresources.Bucket),
		Client: input.clients.S3,
	}
	if err := templateS3.CreateOrFail(); err != nil {
		return microerror.MaskAny(err)
	}

	asgSize := len(input.cluster.Spec.AWS.Workers)

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
		return microerror.MaskAny(err)
	}

	// Calculate big CC checksum, to be referenced in small CC and used in the
	// big CC's filename.
	checksum := sha256.Sum256([]byte(cloudConfig))

	cloudconfigConfig := SmallCloudconfigConfig{
		Filename: s.cloudConfigName(input.prefix, checksum),
		Region:   input.cluster.Spec.AWS.Region,
		S3URI:    s.bucketName(input.cluster),
	}

	cloudconfigS3 := &awsresources.BucketObject{
		Name:   s.cloudConfigRelativePath(input.prefix, checksum),
		Data:   cloudConfig,
		Bucket: input.bucket.(*awsresources.Bucket),
		Client: input.clients.S3,
	}
	if err := cloudconfigS3.CreateOrFail(); err != nil {
		return microerror.MaskAny(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		return microerror.MaskAny(err)
	}

	// Create CloudFormation stack for the ASG.
	stack := awsresources.ASGStack{
		// Dependencies.
		Client: input.clients.CloudFormation,

		// Settings.
		ASGMaxSize:               asgSize,
		ASGMinSize:               asgSize,
		AssociatePublicIPAddress: publicIP,
		AvailabilityZone:         input.cluster.Spec.AWS.AZ,
		ClusterID:                input.cluster.Spec.Cluster.Cluster.ID,
		HealthCheckGracePeriod:   gracePeriodSeconds,
		IAMInstanceProfileName:   input.iamInstanceProfileName,
		ImageID:                  imageID,
		InstanceType:             instanceType,
		KeyName:                  input.keyPairName,
		Name:                     s.asgName(input.cluster, prefixWorker),
		SmallCloudConfig:         smallCloudconfig,
		SubnetID:                 input.subnetID,
		TemplateURL:              templateURL,
		VPCID:                    input.vpcID,
	}

	if err := stack.CreateOrFail(); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *Service) asgName(cluster awstpr.CustomObject, prefix string) string {
	return fmt.Sprintf("%s-%s", cluster.Name, prefix)
}
