package create

import (
	"fmt"

	"github.com/giantswarm/aws-operator/resources"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type launchConfigurationInput struct {
	name                string
	clients             awsutil.Clients
	cluster             awstpr.CustomObject
	tlsAssets           *certificatetpr.CompactTLSAssets
	bucket              resources.Resource
	securityGroup       resources.ResourceWithID
	subnet              *awsresources.Subnet
	keypairName         string
	instanceProfileName string
	prefix              string
}

func (s *Service) createLaunchConfiguration(input launchConfigurationInput) (bool, error) {
	var (
		extension    cloudconfig.Extension
		imageID      string
		instanceType string
	)

	switch input.prefix {
	case prefixMaster:
		extension = NewMasterCloudConfigExtension(input.cluster.Spec, input.tlsAssets)

		// TODO Check only a single master node is provided.
		imageID = input.cluster.Spec.AWS.Masters[0].ImageID
		instanceType = input.cluster.Spec.AWS.Masters[0].InstanceType
	case prefixWorker:
		extension = NewWorkerCloudConfigExtension(input.cluster.Spec, input.tlsAssets)

		// TODO Until multiple worker instance types supported check only a single
		// image ID and instance type is provided.
		imageID = input.cluster.Spec.AWS.Workers[0].ImageID
		instanceType = input.cluster.Spec.AWS.Workers[0].InstanceType
	default:
		return false, microerror.MaskAnyf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.prefix))
	}

	cloudConfigParams := cloudconfig.Params{
		Cluster:   input.cluster.Spec.Cluster,
		Extension: extension,
	}

	cloudConfig, err := s.cloudConfig(input.prefix, cloudConfigParams, input.cluster.Spec, input.tlsAssets)
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	// We now upload the instance cloudconfig to S3 and create a "small
	// cloudconfig" that just fetches the previously uploaded "final
	// cloudconfig" and executes coreos-cloudinit with it as argument.
	// We do this to circumvent the 16KB limit on user-data for EC2 instances.
	cloudconfigConfig := SmallCloudconfigConfig{
		MachineType: input.prefix,
		Region:      input.cluster.Spec.AWS.Region,
		S3DirURI:    s.bucketObjectFullDirPath(input.cluster),
	}

	var cloudconfigS3 resources.Resource
	cloudconfigS3 = &awsresources.BucketObject{
		Name:      s.bucketObjectName(input.cluster, input.prefix),
		Data:      cloudConfig,
		Bucket:    input.bucket.(*awsresources.Bucket),
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
	}
	if err := cloudconfigS3.CreateOrFail(); err != nil {
		return false, microerror.MaskAny(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	securityGroupID, err := input.securityGroup.GetID()
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	var launchConfig *awsresources.LaunchConfiguration
	var launchConfigName string
	var launchConfigCreated bool
	{
		var err error

		launchConfigName, err = launchConfigurationName(input.cluster, input.prefix)
		if err != nil {
			return false, microerror.MaskAny(err)
		}

		launchConfig = &awsresources.LaunchConfiguration{
			Client: input.clients.AutoScaling,
			Name:   launchConfigName,
			IamInstanceProfileName: input.instanceProfileName,
			ImageID:                imageID,
			InstanceType:           instanceType,
			KeyName:                input.keypairName,
			SecurityGroupID:        securityGroupID,
			SmallCloudConfig:       smallCloudconfig,
		}
		launchConfigCreated, err = launchConfig.CreateIfNotExists()
		if err != nil {
			return false, microerror.MaskAny(err)
		}
	}

	return launchConfigCreated, nil
}

func (s *Service) deleteLaunchConfiguration(input launchConfigurationInput) error {
	lc := awsresources.LaunchConfiguration{
		Name: input.name,
	}

	if err := lc.Delete(); err != nil {
		return microerror.MaskAny(err)
	}
	s.logger.Log("debug", fmt.Sprintf("deleted launch configuration '%s'", input.name))

	return nil
}

func launchConfigurationName(cluster awstpr.CustomObject, prefix string) (string, error) {
	if cluster.Spec.Cluster.Cluster.ID == "" {
		return "", microerror.MaskAnyf(missingCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	if prefix == "" {
		return "", microerror.MaskAnyf(missingCloudConfigKeyError, "launchConfiguration prefix")
	}

	return fmt.Sprintf("%s-%s", cluster.Spec.Cluster.Cluster.ID, prefix), nil
}
