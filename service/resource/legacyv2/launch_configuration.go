package legacyv2

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/keyv2"
)

type launchConfigurationInput struct {
	associatePublicIP   bool
	bucket              resources.Resource
	clients             awsutil.Clients
	cluster             v1alpha1.AWSConfig
	instanceProfileName string
	keypairName         string
	name                string
	prefix              string
	securityGroup       resources.ResourceWithID
	ebsStorage          bool
	subnet              *awsresources.Subnet
	tlsAssets           *certificatetpr.CompactTLSAssets
	clusterKeys         *randomkeytpr.CompactRandomKeyAssets
}

func (s *Resource) createLaunchConfiguration(input launchConfigurationInput) (bool, error) {
	var err error
	var imageID string
	var instanceType string
	var template string

	{
		switch input.prefix {
		case prefixMaster:
			imageID = keyv2.MasterImageID(input.cluster)
			instanceType = keyv2.MasterInstanceType(input.cluster)

			template, err = s.cloudConfig.NewMasterTemplate(input.cluster, *input.tlsAssets, *input.clusterKeys)
		case prefixWorker:
			imageID = keyv2.WorkerImageID(input.cluster)
			instanceType = keyv2.WorkerInstanceType(input.cluster)

			template, err = s.cloudConfig.NewWorkerTemplate(input.cluster, *input.tlsAssets)
		default:
			return false, microerror.Maskf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.prefix))
		}

		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	// We now upload the instance cloudconfig to S3 and create a "small
	// cloudconfig" that just fetches the previously uploaded "final
	// cloudconfig" and executes coreos-cloudinit with it as argument.
	// We do this to circumvent the 16KB limit on user-data for EC2 instances.
	cloudconfigConfig := SmallCloudconfigConfig{
		MachineType: input.prefix,
		Region:      input.cluster.Spec.AWS.Region,
		S3URI:       s.bucketName(input.cluster),
	}

	cloudconfigS3 := &awsresources.BucketObject{
		Name:      s.bucketObjectName(input.prefix),
		Data:      template,
		Bucket:    input.bucket.(*awsresources.Bucket),
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
	}
	if err := cloudconfigS3.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		return false, microerror.Mask(err)
	}

	securityGroupID, err := input.securityGroup.GetID()
	if err != nil {
		return false, microerror.Mask(err)
	}

	launchConfigName, err := launchConfigurationName(input.cluster, input.prefix, securityGroupID)
	if err != nil {
		return false, microerror.Mask(err)
	}

	launchConfig := &awsresources.LaunchConfiguration{
		Client: input.clients.AutoScaling,
		Name:   launchConfigName,
		IamInstanceProfileName:   input.instanceProfileName,
		ImageID:                  imageID,
		InstanceType:             instanceType,
		KeyName:                  input.keypairName,
		SecurityGroupID:          securityGroupID,
		SmallCloudConfig:         smallCloudconfig,
		AssociatePublicIpAddress: input.associatePublicIP,
		EBSStorage:               input.ebsStorage,
	}

	launchConfigCreated, err := launchConfig.CreateIfNotExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	return launchConfigCreated, nil
}

func (s *Resource) deleteLaunchConfiguration(input launchConfigurationInput) error {
	groupName := keyv2.SecurityGroupName(input.cluster, input.prefix)
	sg := awsresources.SecurityGroup{
		Description: groupName,
		GroupName:   groupName,
		AWSEntity:   awsresources.AWSEntity{Clients: input.clients},
	}

	sgID, err := sg.GetID()
	if err != nil {
		return microerror.Mask(err)
	}

	workersLCName, err := launchConfigurationName(input.cluster, prefixWorker, sgID)
	if err != nil {
		return microerror.Mask(err)
	}

	lc := awsresources.LaunchConfiguration{
		Client: input.clients.AutoScaling,
		Name:   workersLCName,
	}

	if err := lc.Delete(); err != nil {
		return microerror.Mask(err)
	}
	return nil
}

// launchConfigurationName uses the cluster ID, a prefix and the security group
// ID to produce a launch configuration name.  LC names are their unique
// identifiers in AWS.  The reason we need the securityGroupID in the name is
// that we can only reuse an LC if it has been created for the current SG.
// Otherwise, the SG might not exist anymore.
func launchConfigurationName(cluster v1alpha1.AWSConfig, prefix, securityGroupID string) (string, error) {
	if keyv2.ClusterID(cluster) == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	if prefix == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "launchConfiguration prefix")
	}

	if securityGroupID == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "launchConfiguration securityGroupID")
	}

	return fmt.Sprintf("%s-%s-%s", keyv2.ClusterID(cluster), prefix, securityGroupID), nil
}
