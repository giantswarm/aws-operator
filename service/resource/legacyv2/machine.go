package legacyv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type instanceNameInput struct {
	clusterName string
	prefix      string
	no          int
}

func instanceName(input instanceNameInput) string {
	return fmt.Sprintf(instanceNameFormat, input.clusterName, input.prefix, input.no)
}

type clusterPrefixInput struct {
	clusterName string
	prefix      string
}

func clusterPrefix(input clusterPrefixInput) string {
	return fmt.Sprintf(instanceClusterPrefixFormat, input.clusterName, input.prefix)
}

type runMachinesInput struct {
	clients             awsutil.Clients
	cluster             v1alpha1.AWSConfig
	tlsAssets           *legacy.CompactTLSAssets
	clusterKeys         *randomkeytpr.CompactRandomKeyAssets
	bucket              resources.Resource
	securityGroup       resources.ResourceWithID
	subnet              *awsresources.Subnet
	clusterName         string
	keyPairName         string
	instanceProfileName string
	prefix              string
}

func (s *Resource) runMachines(input runMachinesInput) (bool, []string, error) {
	var (
		anyCreated bool

		machines    []v1alpha1.ClusterNode
		awsMachines []v1alpha1.AWSConfigSpecAWSNode
		instanceIDs []string
	)

	switch input.prefix {
	case prefixMaster:
		machines = input.cluster.Spec.Cluster.Masters
		awsMachines = input.cluster.Spec.AWS.Masters
	case prefixWorker:
		machines = input.cluster.Spec.Cluster.Workers
		awsMachines = input.cluster.Spec.AWS.Workers
	}

	// TODO(nhlfr): Create a separate module for validating specs and execute on the earlier stages.
	if len(machines) != len(awsMachines) {
		return false, nil, microerror.Mask(fmt.Errorf("mismatched number of %s machines in the 'spec' and 'aws' sections: %d != %d",
			input.prefix,
			len(machines),
			len(awsMachines)))
	}

	for i := 0; i < len(machines); i++ {
		name := instanceName(instanceNameInput{
			clusterName: input.clusterName,
			prefix:      input.prefix,
			no:          i,
		})
		created, instanceID, err := s.runMachine(runMachineInput{
			clients:             input.clients,
			cluster:             input.cluster,
			machine:             machines[i],
			awsNode:             awsMachines[i],
			tlsAssets:           input.tlsAssets,
			clusterKeys:         input.clusterKeys,
			bucket:              input.bucket,
			securityGroup:       input.securityGroup,
			subnet:              input.subnet,
			clusterName:         input.clusterName,
			keyPairName:         input.keyPairName,
			instanceProfileName: input.instanceProfileName,
			name:                name,
			prefix:              input.prefix,
		})
		if err != nil {
			return false, nil, microerror.Mask(err)
		}
		if created {
			anyCreated = true
		}

		instanceIDs = append(instanceIDs, instanceID)
	}
	return anyCreated, instanceIDs, nil
}

// if the instance already exists, return (instanceID, false)
// otherwise (nil, true)
func allExistingInstancesMatch(instances *ec2.DescribeInstancesOutput, state awsresources.EC2StateCode) (*string, bool) {
	// If the instance doesn't exist, then the Reservations field should be nil.
	// Otherwise, it will contain a slice of instances (which is going to contain our one instance we queried for).
	// TODO(nhlfr): Check whether the instance has correct parameters. That will be most probably done when we
	// will introduce the interface for creating, deleting and updating resources.
	if instances.Reservations != nil {
		for _, r := range instances.Reservations {
			for _, i := range r.Instances {
				if *i.State.Code != int64(state) {
					return i.InstanceId, false
				}
			}
		}
	}
	return nil, true
}

type runMachineInput struct {
	clients             awsutil.Clients
	cluster             v1alpha1.AWSConfig
	machine             v1alpha1.ClusterNode
	awsNode             v1alpha1.AWSConfigSpecAWSNode
	tlsAssets           *legacy.CompactTLSAssets
	clusterKeys         *randomkeytpr.CompactRandomKeyAssets
	bucket              resources.Resource
	securityGroup       resources.ResourceWithID
	subnet              *awsresources.Subnet
	clusterName         string
	keyPairName         string
	instanceProfileName string
	name                string
	prefix              string
}

func (s *Resource) runMachine(input runMachineInput) (bool, string, error) {
	var template string
	var err error
	{
		switch input.prefix {
		case prefixMaster:
			template, err = s.cloudConfig.NewMasterTemplate(input.cluster, *input.tlsAssets, *input.clusterKeys)
		case prefixWorker:
			template, err = s.cloudConfig.NewWorkerTemplate(input.cluster, *input.tlsAssets)
		default:
			return false, "", microerror.Maskf(invalidCloudconfigExtensionNameError, fmt.Sprintf("Invalid extension name '%s'", input.prefix))
		}

		if err != nil {
			return false, "", microerror.Mask(err)
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

	var cloudconfigS3 resources.Resource
	cloudconfigS3 = &awsresources.BucketObject{
		Name:      s.bucketObjectName(input.prefix),
		Data:      template,
		Bucket:    input.bucket.(*awsresources.Bucket),
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
	}
	if err := cloudconfigS3.CreateOrFail(); err != nil {
		return false, "", microerror.Mask(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		return false, "", microerror.Mask(err)
	}

	securityGroupID, err := input.securityGroup.GetID()
	if err != nil {
		return false, "", microerror.Mask(err)
	}

	subnetID, err := input.subnet.GetID()
	if err != nil {
		return false, "", microerror.Mask(err)
	}

	var instance *awsresources.Instance
	var instanceCreated bool
	{
		var err error
		instance = &awsresources.Instance{
			Name:                   input.name,
			ClusterName:            input.clusterName,
			ImageID:                input.awsNode.ImageID,
			InstanceType:           input.awsNode.InstanceType,
			KeyName:                input.keyPairName,
			MinCount:               1,
			MaxCount:               1,
			SmallCloudconfig:       smallCloudconfig,
			IamInstanceProfileName: input.instanceProfileName,
			PlacementAZ:            input.cluster.Spec.AWS.AZ,
			SecurityGroupID:        securityGroupID,
			SubnetID:               subnetID,
			Logger:                 s.logger,
			AWSEntity:              awsresources.AWSEntity{Clients: input.clients},
		}
		instanceCreated, err = instance.CreateIfNotExists()
		if err != nil {
			return false, "", microerror.Mask(err)
		}
	}

	if instanceCreated {
		s.logger.Log("info", fmt.Sprintf("instance '%s' reserved", input.name))
	} else {
		s.logger.Log("info", fmt.Sprintf("instance '%s' already exists, reusing", input.name))
	}

	s.logger.Log("info", fmt.Sprintf("instance '%s' tagged", input.name))

	return instanceCreated, instance.ID(), nil
}

type deleteMachinesInput struct {
	clients     awsutil.Clients
	spec        v1alpha1.AWSConfigSpec
	clusterName string
	prefix      string
}

func (s *Resource) deleteMachines(input deleteMachinesInput) error {
	pattern := clusterPrefix(clusterPrefixInput{
		clusterName: input.clusterName,
		prefix:      input.prefix,
	})
	instances, err := awsresources.FindInstances(awsresources.FindInstancesInput{
		Clients: input.clients,
		Logger:  s.logger,
		Pattern: pattern,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, instance := range instances {
		if err := instance.Delete(); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

type deleteMachineInput struct {
	name    string
	clients awsutil.Clients
	machine v1alpha1.AWSConfigSpecAWSNode
}

func validateIDs(ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if id == "" {
			return false
		}
	}

	return true
}
