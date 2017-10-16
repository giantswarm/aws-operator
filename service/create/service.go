package create

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/awstpr"
	awsinfo "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfig"
	"github.com/giantswarm/aws-operator/service/key"
)

const (
	// The format of instance's name is "[name of cluster]-[prefix ('master' or 'worker')]-[number]".
	instanceNameFormat string = "%s-%s-%d"
	// The format of prefix inside a cluster "[name of cluster]-[prefix ('master' or 'worker')]".
	instanceClusterPrefixFormat string = "%s-%s"
	// Prefixes used for machine names.
	prefixMaster  string = "master"
	prefixWorker  string = "worker"
	prefixIngress string = "ingress"
	// Suffixes used for subnets
	suffixPublic  string = "public"
	suffixPrivate string = "private"
	// Number of retries of RunInstances to wait for Roles to propagate to
	// Instance Profiles
	runInstancesRetries = 10
	// The number of seconds AWS will wait, before issuing a health check on
	// instances in an Auto Scaling Group.
	gracePeriodSeconds = 10
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Dependencies.
	CertWatcher *certificatetpr.Service
	CloudConfig *cloudconfig.CloudConfig
	K8sClient   kubernetes.Interface
	KeyWatcher  *randomkeytpr.Service
	Logger      micrologger.Logger

	// Settings.
	AwsConfig     awsutil.Config
	AwsHostConfig awsutil.Config
	PubKeyFile    string
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		CloudConfig: nil,
		K8sClient:   nil,
		KeyWatcher:  nil,
		Logger:      nil,

		// Settings.
		AwsConfig:     awsutil.Config{},
		AwsHostConfig: awsutil.Config{},
		PubKeyFile:    "",
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.KeyWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.KeyWatcher must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsConfig must not be empty")
	}
	if config.AwsHostConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsHostConfig must not be empty")
	}
	if config.PubKeyFile == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.PubKeyFile must not be empty")
	}

	var err error

	var newTPR *tpr.TPR
	{
		tprConfig := tpr.DefaultConfig()

		tprConfig.K8sClient = config.K8sClient
		tprConfig.Logger = config.Logger

		tprConfig.Description = awstpr.Description
		tprConfig.Name = awstpr.Name
		tprConfig.Version = awstpr.VersionV1

		newTPR, err = tpr.New(tprConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Dependencies.
		certWatcher: config.CertWatcher,
		cloudConfig: config.CloudConfig,
		k8sClient:   config.K8sClient,
		keyWatcher:  config.KeyWatcher,
		logger:      config.Logger,

		// Internals
		bootOnce: sync.Once{},
		tpr:      newTPR,

		// Settings.
		awsConfig:     config.AwsConfig,
		awsHostConfig: config.AwsHostConfig,
		pubKeyFile:    config.PubKeyFile,
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	certWatcher *certificatetpr.Service
	cloudConfig *cloudconfig.CloudConfig
	k8sClient   kubernetes.Interface
	keyWatcher  *randomkeytpr.Service
	logger      micrologger.Logger

	// Internals.
	bootOnce sync.Once
	tpr      *tpr.TPR

	// Settings.
	awsConfig     awsutil.Config
	awsHostConfig awsutil.Config
	pubKeyFile    string
}

type Event struct {
	Type   string
	Object *awstpr.CustomObject
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		err := s.tpr.CreateAndWait()
		if tpr.IsAlreadyExists(err) {
			s.logger.Log("debug", "third party resource already exists")
		} else if err != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", err))
			return
		}

		s.logger.Log("debug", "starting list/watch")

		newResourceEventHandler := &cache.ResourceEventHandlerFuncs{
			AddFunc:    s.addFunc,
			DeleteFunc: s.deleteFunc,
			UpdateFunc: s.updateFunc,
		}
		newZeroObjectFactory := &tpr.ZeroObjectFactoryFuncs{
			NewObjectFunc:     func() runtime.Object { return &awstpr.CustomObject{} },
			NewObjectListFunc: func() runtime.Object { return &awstpr.List{} },
		}

		s.tpr.NewInformer(newResourceEventHandler, newZeroObjectFactory).Run(nil)
	})
}

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
	cluster             awstpr.CustomObject
	tlsAssets           *certificatetpr.CompactTLSAssets
	clusterKeys         *randomkeytpr.CompactRandomKeyAssets
	bucket              resources.Resource
	securityGroup       resources.ResourceWithID
	subnet              *awsresources.Subnet
	clusterName         string
	keyPairName         string
	instanceProfileName string
	prefix              string
}

func (s *Service) runMachines(input runMachinesInput) (bool, []string, error) {
	var (
		anyCreated bool

		machines    []spec.Node
		awsMachines []awsinfo.Node
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

func (s *Service) uploadCloudconfigToS3(svc *s3.S3, s3Bucket, path, data string) error {
	if _, err := svc.PutObject(&s3.PutObjectInput{
		Body:          strings.NewReader(data),
		Bucket:        aws.String(s3Bucket),
		Key:           aws.String(path),
		ContentLength: aws.Int64(int64(len(data))),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

type runMachineInput struct {
	clients             awsutil.Clients
	cluster             awstpr.CustomObject
	machine             spec.Node
	awsNode             awsinfo.Node
	tlsAssets           *certificatetpr.CompactTLSAssets
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

func (s *Service) runMachine(input runMachineInput) (bool, string, error) {
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
	spec        awstpr.Spec
	clusterName string
	prefix      string
}

func (s *Service) deleteMachines(input deleteMachinesInput) error {
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
	machine spec.Node
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

func (s *Service) addFunc(obj interface{}) {
	customObject, ok := obj.(*awstpr.CustomObject)
	if !ok {
		s.logger.Log("error", "could not convert to awstpr.CustomObject")
		return
	}
	cluster := *customObject

	s.logger.Log("info", fmt.Sprintf("creating cluster '%s'", key.ClusterID(cluster)))

	if err := validateCluster(cluster); err != nil {
		s.logger.Log("error", "cluster spec is invalid: '%#v'", err)
		return
	}

	err := s.processCluster(cluster)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("error processing cluster '%s': '%#v'", key.ClusterID(cluster), err))
		return
	}

	s.logger.Log("info", fmt.Sprintf("cluster '%s' processed", key.ClusterID(cluster)))
}

func (s *Service) processCluster(cluster awstpr.CustomObject) error {
	// Create cluster namespace in k8s.
	if err := s.createClusterNamespace(cluster.Spec.Cluster); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create cluster namespace: '%#v'", err))
	}

	// Create AWS guest cluster client.
	s.awsConfig.Region = cluster.Spec.AWS.Region
	clients := awsutil.NewClients(s.awsConfig)
	if err := s.awsConfig.SetAccountID(clients.IAM); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not retrieve guest amazon account id: '%#v'", err))
	}

	// Create AWS host cluster client.
	s.awsHostConfig.Region = cluster.Spec.AWS.Region
	hostClients := awsutil.NewClients(s.awsHostConfig)
	if err := s.awsHostConfig.SetAccountID(hostClients.IAM); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not retrieve host amazon account id: '%#v'", err))
	}

	// An EC2 Keypair is needed for legacy clusters. New clusters provide SSH keys via cloud config.
	if !key.HasClusterVersion(cluster) {
		// Create keypair.
		var keyPair resources.ReusableResource
		var keyPairCreated bool
		{
			var err error
			keyPair = &awsresources.KeyPair{
				ClusterName: key.ClusterID(cluster),
				Provider:    awsresources.NewFSKeyPairProvider(s.pubKeyFile),
				AWSEntity:   awsresources.AWSEntity{Clients: clients},
			}
			keyPairCreated, err = keyPair.CreateIfNotExists()
			if err != nil {
				return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create keypair: '%#v'", err))
			}
		}

		if keyPairCreated {
			s.logger.Log("info", fmt.Sprintf("created keypair '%s'", key.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("keypair '%s' already exists, reusing", key.ClusterID(cluster)))
		}
	}

	s.logger.Log("info", fmt.Sprintf("waiting for k8s secrets..."))
	clusterID := key.ClusterID(cluster)
	certs, err := s.certWatcher.SearchCerts(clusterID)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get certificates from secrets: '%#v'", err))
	}

	// Create Encryption key
	s.logger.Log("info", fmt.Sprintf("waiting for k8s keys..."))
	keys, err := s.keyWatcher.SearchKeys(clusterID)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get keys from secrets: '%#v'", err))
	}

	// Create KMS key.
	kmsKey := &awsresources.KMSKey{
		Name:      key.ClusterID(cluster),
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}

	kmsCreated, kmsKeyErr := kmsKey.CreateIfNotExists()
	if kmsKeyErr != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create KMS key: '%#v'", kmsKeyErr))
	}

	if kmsCreated {
		s.logger.Log("info", fmt.Sprintf("created KMS key for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("kms key '%s' already exists, reusing", kmsKey.Name))
	}

	// Encode TLS assets.
	tlsAssets, err := s.encodeTLSAssets(certs, clients.KMS, kmsKey.Arn())
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not encode TLS assets: '%#v'", err))
	}

	// Encode Key assets.
	clusterKeys, err := s.encodeKeyAssets(keys, clients.KMS, kmsKey.Arn())
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not encode Keys assets: '%#v'", err))
	}

	bucketName := s.bucketName(cluster)

	// Create master IAM policy.
	var masterPolicy resources.NamedResource
	var masterPolicyCreated bool
	{
		var err error
		masterPolicy = &awsresources.Policy{
			ClusterID:  key.ClusterID(cluster),
			KMSKeyArn:  kmsKey.Arn(),
			PolicyType: prefixMaster,
			S3Bucket:   bucketName,
			AWSEntity:  awsresources.AWSEntity{Clients: clients},
		}
		masterPolicyCreated, err = masterPolicy.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create master policy: '%#v'", err))
		}
	}
	if masterPolicyCreated {
		s.logger.Log("info", fmt.Sprintf("created master policy for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("master policy for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
	}

	// Create worker IAM policy.
	var workerPolicy resources.NamedResource
	var workerPolicyCreated bool
	{
		var err error
		workerPolicy = &awsresources.Policy{
			ClusterID:  key.ClusterID(cluster),
			KMSKeyArn:  kmsKey.Arn(),
			PolicyType: prefixWorker,
			S3Bucket:   bucketName,
			AWSEntity:  awsresources.AWSEntity{Clients: clients},
		}
		workerPolicyCreated, err = workerPolicy.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create worker policy: '%#v'", err))
		}
	}
	if workerPolicyCreated {
		s.logger.Log("info", fmt.Sprintf("created worker policy for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("worker policy for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
	}

	// Create S3 bucket.
	var bucket resources.ReusableResource
	var bucketCreated bool
	{
		var err error
		bucket = &awsresources.Bucket{
			Name:      bucketName,
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}
		bucketCreated, err = bucket.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create S3 bucket: '%#v'", err))
		}
	}

	if bucketCreated {
		s.logger.Log("info", fmt.Sprintf("created bucket '%s'", bucketName))
	} else {
		s.logger.Log("info", fmt.Sprintf("bucket '%s' already exists, reusing", bucketName))
	}

	// Create VPC.
	var vpc resources.ResourceWithID
	vpc = &awsresources.VPC{
		CidrBlock: cluster.Spec.AWS.VPC.CIDR,
		Name:      key.ClusterID(cluster),
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	vpcCreated, err := vpc.CreateIfNotExists()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create VPC: '%#v'", err))
	}
	if vpcCreated {
		s.logger.Log("info", fmt.Sprintf("created vpc for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("vpc for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
	}
	vpcID, err := vpc.GetID()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get VPC ID: '%#v'", err))
	}

	// Create VPC peering connection.
	vpcPeeringConection := &awsresources.VPCPeeringConnection{
		VPCId:     vpcID,
		PeerVPCId: cluster.Spec.AWS.VPC.PeerID,
		AWSEntity: awsresources.AWSEntity{
			Clients:     clients,
			HostClients: hostClients,
		},
	}
	vpcPeeringConnectionCreated, err := vpcPeeringConection.CreateIfNotExists()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create vpc peering connection: '%#v'", err))
	}
	if vpcPeeringConnectionCreated {
		s.logger.Log("info", fmt.Sprintf("created vpc peering connection for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("vpc peering connection for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
	}

	conn, err := vpcPeeringConection.FindExisting()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not find vpc peering connection: '%#v'", err))
	}

	// Create internet gateway.
	var internetGateway resources.ResourceWithID
	internetGateway = &awsresources.InternetGateway{
		Name:  key.ClusterID(cluster),
		VpcID: vpcID,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	internetGatewayCreated, err := internetGateway.CreateIfNotExists()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create internet gateway: '%#v'", err))
	}
	if internetGatewayCreated {
		s.logger.Log("info", fmt.Sprintf("created internet gateway for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("internet gateway for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
	}

	// Create masters security group.
	mastersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixMaster),
		VPCID:     vpcID,
	}
	mastersSecurityGroup, err := s.createSecurityGroup(mastersSGInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create security group '%s': '%#v'", mastersSGInput.GroupName, err))
	}
	mastersSecurityGroupID, err := mastersSecurityGroup.GetID()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get security group '%s' ID: '%#v'", mastersSGInput.GroupName, err))
	}

	// Create workers security group.
	workersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixWorker),
		VPCID:     vpcID,
	}
	workersSecurityGroup, err := s.createSecurityGroup(workersSGInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create security group '%s': '%#v'", workersSGInput.GroupName, err))
	}
	workersSecurityGroupID, err := workersSecurityGroup.GetID()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get security group '%s' ID: '%#v'", workersSGInput.GroupName, err))
	}

	// Create ingress ELB security group.
	ingressSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixIngress),
		VPCID:     vpcID,
	}
	ingressSecurityGroup, err := s.createSecurityGroup(ingressSGInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create security group '%s': '%#v'", ingressSGInput.GroupName, err))
	}
	ingressSecurityGroupID, err := ingressSecurityGroup.GetID()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get security group '%s' ID: '%#v'", ingressSGInput.GroupName, err))
	}

	// Create rules for the security groups.
	rulesInput := rulesInput{
		Cluster:                cluster,
		MastersSecurityGroupID: mastersSecurityGroupID,
		WorkersSecurityGroupID: workersSecurityGroupID,
		IngressSecurityGroupID: ingressSecurityGroupID,
		HostClusterCIDR:        *conn.AccepterVpcInfo.CidrBlock,
	}

	if err := mastersSecurityGroup.ApplyRules(rulesInput.masterRules()); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create rules for security group '%s': '%#v", mastersSecurityGroup.GroupName, err))
	}

	if err := workersSecurityGroup.ApplyRules(rulesInput.workerRules()); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create rules for security group '%s': '%#v'", workersSecurityGroup.GroupName, err))
	}

	if err := ingressSecurityGroup.ApplyRules(rulesInput.ingressRules()); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create rules for security group '%s': '%#v'", ingressSecurityGroup.GroupName, err))
	}

	// Create route table.
	routeTable := &awsresources.RouteTable{
		Name:   key.ClusterID(cluster),
		VpcID:  vpcID,
		Client: clients.EC2,
	}
	routeTableCreated, err := routeTable.CreateIfNotExists()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create route table: '%#v'", err))
	}
	if routeTableCreated {
		s.logger.Log("info", "created route table")
	} else {
		s.logger.Log("info", "route table already exists, reusing")
	}

	if err := routeTable.MakePublic(); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not make route table public: '%#v'", err))
	}

	// Create public subnet.
	subnetInput := SubnetInput{
		Name:       key.SubnetName(cluster, suffixPublic),
		CidrBlock:  cluster.Spec.AWS.VPC.PublicSubnetCIDR,
		Clients:    clients,
		Cluster:    cluster,
		MakePublic: true,
		RouteTable: routeTable,
		VpcID:      vpcID,
	}
	publicSubnet, err := s.createSubnet(subnetInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create public subnet: '%#v'", err))
	}

	publicSubnetID, err := publicSubnet.GetID()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get public subnet ID: '%#v'", err))
	}

	var privateSubnet *awsresources.Subnet

	if key.HasClusterVersion(cluster) {
		// Create private subnet.
		subnetInput := SubnetInput{
			Name:       key.SubnetName(cluster, suffixPrivate),
			CidrBlock:  cluster.Spec.AWS.VPC.PrivateSubnetCIDR,
			Clients:    clients,
			Cluster:    cluster,
			MakePublic: false,
			VpcID:      vpcID,
		}
		privateSubnet, err = s.createSubnet(subnetInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create public subnet: '%#v'", err))
		}

		_, err := privateSubnet.GetID()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get private subnet ID: '%#v'", err))
		}
	}

	publicRoute := &awsresources.Route{
		RouteTable:           *routeTable,
		DestinationCidrBlock: *conn.AccepterVpcInfo.CidrBlock,
		VpcID:                *conn.VpcPeeringConnectionId,
		AWSEntity:            awsresources.AWSEntity{Clients: clients},
	}

	publicRouteCreated, err := publicRoute.CreateIfNotExists()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not add host vpc route: '%#v'", err))
	}
	if publicRouteCreated {
		s.logger.Log("info", fmt.Sprintf("created host vpc route for cluster '%s'", key.ClusterID(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("host vpc route for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
	}

	for _, privateRouteTableName := range cluster.Spec.AWS.VPC.RouteTableNames {
		privateRouteTable := &awsresources.RouteTable{
			Name:   privateRouteTableName,
			VpcID:  cluster.Spec.AWS.VPC.PeerID,
			Client: hostClients.EC2,
		}

		privateRoute := &awsresources.Route{
			RouteTable:           *privateRouteTable,
			DestinationCidrBlock: *conn.RequesterVpcInfo.CidrBlock,
			VpcID:                *conn.VpcPeeringConnectionId,
			AWSEntity:            awsresources.AWSEntity{Clients: hostClients},
		}

		privateRouteCreated, err := privateRoute.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not add guest vpc route: '%#v'", err))
		}
		if privateRouteCreated {
			s.logger.Log("info", fmt.Sprintf("created guest vpc route for cluster '%s'", key.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("host vpc guest for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
		}

	}

	if key.HasClusterVersion(cluster) {
		// Create NAT gateway.
		var natGateway resources.ResourceWithID
		natGateway = &awsresources.NatGateway{
			Name:   key.ClusterID(cluster),
			Subnet: publicSubnet,
			// Dependencies.
			Logger:    s.logger,
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}
		natGatewayCreated, err := natGateway.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create nat gateway: '%#v'", err))
		}
		if natGatewayCreated {
			s.logger.Log("info", fmt.Sprintf("created nat gateway for cluster '%s'", key.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("nat gateway for cluster '%s' already exists, reusing", key.ClusterID(cluster)))
		}
	}

	mastersInput := runMachinesInput{
		clients:             clients,
		cluster:             cluster,
		tlsAssets:           tlsAssets,
		clusterKeys:         clusterKeys,
		clusterName:         key.ClusterID(cluster),
		bucket:              bucket,
		securityGroup:       mastersSecurityGroup,
		subnet:              publicSubnet,
		instanceProfileName: masterPolicy.GetName(),
		prefix:              prefixMaster,
	}

	// An EC2 Keypair is needed for legacy clusters. New clusters provide SSH keys via cloud config.
	if !key.HasClusterVersion(cluster) {
		mastersInput.keyPairName = key.ClusterID(cluster)
	}

	// Run masters.
	anyMastersCreated, masterIDs, err := s.runMachines(mastersInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not start masters: '%#v'", err))
	}

	if !validateIDs(masterIDs) {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("master nodes had invalid instance IDs: %v", masterIDs))
	}

	// Create apiserver load balancer.
	lbInput := LoadBalancerInput{
		Name:        cluster.Spec.Cluster.Kubernetes.API.Domain,
		Clients:     clients,
		Cluster:     cluster,
		InstanceIDs: masterIDs,
		PortsToOpen: awsresources.PortPairs{
			{
				PortELB:      cluster.Spec.Cluster.Kubernetes.API.SecurePort,
				PortInstance: cluster.Spec.Cluster.Kubernetes.API.SecurePort,
			},
		},
		SecurityGroupID: mastersSecurityGroupID,
		SubnetID:        publicSubnetID,
	}

	apiLB, err := s.createLoadBalancer(lbInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create apiserver load balancer: '%#v'", err))
	}

	// Assign the ProxyProtocol policy to the apiserver load balancer.
	if err := apiLB.AssignProxyProtocolPolicy(); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not assign proxy protocol policy: '%#v'", err))
	}

	// Create etcd load balancer.
	lbInput = LoadBalancerInput{
		Name:        cluster.Spec.Cluster.Etcd.Domain,
		Clients:     clients,
		Cluster:     cluster,
		InstanceIDs: masterIDs,
		PortsToOpen: awsresources.PortPairs{
			{
				PortELB:      httpsPort,
				PortInstance: cluster.Spec.Cluster.Etcd.Port,
			},
		},
		SecurityGroupID: mastersSecurityGroupID,
		SubnetID:        publicSubnetID,
	}

	if key.HasClusterVersion(cluster) {
		lbInput.Scheme = "internal"
	}

	etcdLB, err := s.createLoadBalancer(lbInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create etcd load balancer: '%#v'", err))
	}

	// Masters were created but the master IAM policy existed from a previous
	// execution. Its likely that previous execution failed. IAM policies can't
	// be reused for EC2 instances.
	if anyMastersCreated && !masterPolicyCreated {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("cluster '%s' cannot be processed. As IAM policy for master nodes cannot be reused. Please delete this cluster.", key.ClusterID(cluster)))
	}

	// Create Ingress load balancer.
	lbInput = LoadBalancerInput{
		Name:    cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
		Clients: clients,
		Cluster: cluster,
		PortsToOpen: awsresources.PortPairs{
			{
				PortELB:      httpsPort,
				PortInstance: cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
			},
			{
				PortELB:      httpPort,
				PortInstance: cluster.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
			},
		},
		SecurityGroupID: ingressSecurityGroupID,
		SubnetID:        publicSubnetID,
	}

	ingressLB, err := s.createLoadBalancer(lbInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create ingress load balancer: '%#v'", err))
	}

	// Assign the ProxyProtocol policy to the Ingress load balancer.
	if err := ingressLB.AssignProxyProtocolPolicy(); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not assign proxy protocol policy: '%#v'", err))
	}

	s.logger.Log("info", fmt.Sprintf("created ingress load balancer"))

	// Create a launch configuration for the worker nodes.
	lcInput := launchConfigurationInput{
		clients:             clients,
		cluster:             cluster,
		clusterKeys:         clusterKeys,
		tlsAssets:           tlsAssets,
		bucket:              bucket,
		securityGroup:       workersSecurityGroup,
		subnet:              publicSubnet,
		instanceProfileName: workerPolicy.GetName(),
		prefix:              prefixWorker,
		ebsStorage:          true,
	}

	// An EC2 Keypair is needed for legacy clusters. New clusters provide SSH keys via cloud config.
	if !key.HasClusterVersion(cluster) {
		lcInput.keypairName = key.ClusterID(cluster)
	}

	lcCreated, err := s.createLaunchConfiguration(lcInput)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create launch config: '%#v'", err))
	}

	if lcCreated {
		s.logger.Log("info", fmt.Sprintf("created worker launch config"))
	} else {
		s.logger.Log("info", fmt.Sprintf("launch config %s already exists, reusing", key.ClusterID(cluster)))
	}

	workersLCName, err := launchConfigurationName(cluster, "worker", workersSecurityGroupID)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get launch config name: '%#v'", err))
	}

	asg := awsresources.AutoScalingGroup{
		Client:                  clients.AutoScaling,
		Name:                    key.AutoScalingGroupName(cluster, prefixWorker),
		ClusterID:               key.ClusterID(cluster),
		MinSize:                 key.WorkerCount(cluster),
		MaxSize:                 key.WorkerCount(cluster),
		AvailabilityZone:        key.AvailabilityZone(cluster),
		LaunchConfigurationName: workersLCName,
		LoadBalancerName:        ingressLB.Name,
		VPCZoneIdentifier:       publicSubnetID,
		HealthCheckGracePeriod:  gracePeriodSeconds,
	}

	asgCreated, err := asg.CreateIfNotExists()
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create auto scaling group: '%#v'", err))
	}

	if asgCreated {
		s.logger.Log("info", fmt.Sprintf("created auto scaling group '%s' with size %v", asg.Name, key.WorkerCount(cluster)))
	} else {
		s.logger.Log("info", fmt.Sprintf("auto scaling group '%s' already exists, reusing", asg.Name))
	}

	// Create Record Sets for the Load Balancers.
	recordSetInputs := []recordSetInput{
		{
			Cluster:      cluster,
			Client:       clients.Route53,
			Resource:     apiLB,
			Domain:       cluster.Spec.Cluster.Kubernetes.API.Domain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.API,
			Type:         route53.RRTypeA,
		},
		{
			Cluster:      cluster,
			Client:       clients.Route53,
			Resource:     etcdLB,
			Domain:       cluster.Spec.Cluster.Etcd.Domain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.Etcd,
			Type:         route53.RRTypeA,
		},
		{
			Cluster:      cluster,
			Client:       clients.Route53,
			Resource:     ingressLB,
			Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.Ingress,
			Type:         route53.RRTypeA,
		},
		{
			Cluster:      cluster,
			Client:       clients.Route53,
			Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.WildcardDomain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.Ingress,
			Value:        cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
			Type:         route53.RRTypeCname,
		},
	}

	var rsErr error
	for _, input := range recordSetInputs {
		if rsErr = s.createRecordSet(input); rsErr != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create record set '%#v'", rsErr))
		}
	}
	if rsErr == nil {
		s.logger.Log("info", fmt.Sprintf("created DNS records for load balancers"))
	}

	return nil
}

func (s *Service) deleteFunc(obj interface{}) {
	// TODO(nhlfr): Move this to a separate operator.

	// We can receive an instance of awstpr.CustomObject or cache.DeletedFinalStateUnknown.
	// We need to assert the type properly and log an error when we cannot do that.
	// Also, the cache.DeleteFinalStateUnknown object can contain the proper CustomObject,
	// but doesn't have to.
	// https://github.com/kubernetes/client-go/blob/7ba6be594966f4bec08a57a60c855d17a9f7000a/tools/cache/delta_fifo.go#L674-L677
	var cluster awstpr.CustomObject
	clusterPtr, ok := obj.(*awstpr.CustomObject)
	if ok {
		cluster = *clusterPtr
	} else {
		deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			s.logger.Log("error", "received unknown type of third-party object")
			return
		}
		clusterPtr, ok := deletedObj.Obj.(*awstpr.CustomObject)
		if !ok {
			s.logger.Log("error", "received the proper delete request, but the type of third-party object is unknown")
			return
		}
		cluster = *clusterPtr
	}

	if err := validateCluster(cluster); err != nil {
		s.logger.Log("error", "cluster spec is invalid: '%#v'", err)
		return
	}

	if err := s.deleteClusterNamespace(cluster.Spec.Cluster); err != nil {
		s.logger.Log("error", "could not delete cluster namespace:", err)
	}

	clients := awsutil.NewClients(s.awsConfig)
	err := s.awsConfig.SetAccountID(clients.IAM)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve amazon account id: '%#v'", err))
		return
	}

	// Retrieve AWS host cluster client.
	s.awsHostConfig.Region = cluster.Spec.AWS.Region
	hostClients := awsutil.NewClients(s.awsHostConfig)
	if err := s.awsHostConfig.SetAccountID(hostClients.IAM); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve host amazon account id: '%#v'", err))
	}

	// Delete masters.
	s.logger.Log("info", "deleting masters...")
	if err := s.deleteMachines(deleteMachinesInput{
		clients:     clients,
		clusterName: key.ClusterID(cluster),
		prefix:      prefixMaster,
	}); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", "deleted masters")
	}

	// Delete workers Auto Scaling Group.
	asg := awsresources.AutoScalingGroup{
		Client: clients.AutoScaling,
		Name:   key.AutoScalingGroupName(cluster, prefixWorker),
	}

	if err := asg.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", "deleted workers auto scaling group")
	}

	// Delete workers launch configuration.
	lcInput := launchConfigurationInput{
		clients: clients,
		cluster: cluster,
		prefix:  "worker",
	}
	if err := s.deleteLaunchConfiguration(lcInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", "deleted worker launch config")
	}

	// Delete Record Sets.
	apiLBName, err := loadBalancerName(cluster.Spec.Cluster.Kubernetes.API.Domain, cluster)
	etcdLBName, err := loadBalancerName(cluster.Spec.Cluster.Etcd.Domain, cluster)
	ingressLBName, err := loadBalancerName(cluster.Spec.Cluster.Kubernetes.IngressController.Domain, cluster)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		apiLB, err := awsresources.NewELBFromExisting(apiLBName, clients.ELB)
		etcdLB, err := awsresources.NewELBFromExisting(etcdLBName, clients.ELB)
		ingressLB, err := awsresources.NewELBFromExisting(ingressLBName, clients.ELB)
		if err != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", err))
		} else {
			recordSetInputs := []recordSetInput{
				{
					Cluster:      cluster,
					Client:       clients.Route53,
					Resource:     apiLB,
					Domain:       cluster.Spec.Cluster.Kubernetes.API.Domain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.API,
					Type:         route53.RRTypeA,
				},
				{
					Cluster:      cluster,
					Client:       clients.Route53,
					Resource:     etcdLB,
					Domain:       cluster.Spec.Cluster.Etcd.Domain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.Etcd,
					Type:         route53.RRTypeA,
				},
				{
					Cluster:      cluster,
					Client:       clients.Route53,
					Resource:     ingressLB,
					Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.Ingress,
					Type:         route53.RRTypeA,
				},
				{
					Cluster:      cluster,
					Client:       clients.Route53,
					Value:        cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
					Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.WildcardDomain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.Ingress,
					Type:         route53.RRTypeCname,
				},
			}

			var rsErr error
			for _, input := range recordSetInputs {
				if rsErr = s.deleteRecordSet(input); rsErr != nil {
					s.logger.Log("error", fmt.Sprintf("%#v", rsErr))
				}
			}
			if rsErr == nil {
				s.logger.Log("info", "deleted record sets")
			}
		}
	}

	// Delete Load Balancers.
	loadBalancerInputs := []LoadBalancerInput{
		{
			Name:    cluster.Spec.Cluster.Kubernetes.API.Domain,
			Clients: clients,
			Cluster: cluster,
		},
		{
			Name:    cluster.Spec.Cluster.Etcd.Domain,
			Clients: clients,
			Cluster: cluster,
		},
		{
			Name:    cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
			Clients: clients,
			Cluster: cluster,
		},
	}

	var elbErr error
	for _, lbInput := range loadBalancerInputs {
		if elbErr = s.deleteLoadBalancer(lbInput); elbErr != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", elbErr))
		}
	}
	if elbErr == nil {
		s.logger.Log("info", "deleted ELBs")
	}

	// Delete route table.
	var routeTable resources.ResourceWithID
	routeTable = &awsresources.RouteTable{
		Name:   key.ClusterID(cluster),
		Client: clients.EC2,
	}
	if err := routeTable.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete route table: '%#v'", err))
	} else {
		s.logger.Log("info", "deleted route table")
	}

	// Sync VPC.
	var vpc resources.ResourceWithID
	vpc = &awsresources.VPC{
		Name:      key.ClusterID(cluster),
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	vpcID, err := vpc.GetID()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	}

	if key.HasClusterVersion(cluster) {
		// Delete NAT gateway.
		natGateway := &awsresources.NatGateway{
			Name: key.ClusterID(cluster),
			// Dependencies.
			Logger:    s.logger,
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}
		if err := natGateway.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete nat gateway: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted nat gateway")
		}
	}

	// Delete internet gateway.
	internetGateway := &awsresources.InternetGateway{
		Name:  key.ClusterID(cluster),
		VpcID: vpcID,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	if err := internetGateway.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete internet gateway: '%#v'", err))
	} else {
		s.logger.Log("info", "deleted internet gateway")
	}

	// Delete public subnet.
	subnetInput := SubnetInput{
		Name:    key.SubnetName(cluster, suffixPublic),
		Clients: clients,
	}
	if err := s.deleteSubnet(subnetInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete public subnet: '%#v'", err))
	} else {
		s.logger.Log("info", "deleted public subnet")
	}

	// Delete private subnet for new clusters. Legacy clusters only have public subnets.
	if key.HasClusterVersion(cluster) {
		subnetInput = SubnetInput{
			Name:    key.SubnetName(cluster, suffixPrivate),
			Clients: clients,
		}
		if err := s.deleteSubnet(subnetInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete private subnet: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted private subnet")
		}
	}

	// Before the security groups can be deleted any rules referencing other
	// groups must first be deleted.
	mastersSGRulesInput := securityGroupRulesInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixMaster),
	}
	if err := s.deleteSecurityGroupRules(mastersSGRulesInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete rules for security group '%s': '%#v'", mastersSGRulesInput.GroupName, err))
	}

	workersSGRulesInput := securityGroupRulesInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixWorker),
	}
	if err := s.deleteSecurityGroupRules(workersSGRulesInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete rules for security group '%s': '%#v'", mastersSGRulesInput.GroupName, err))
	}

	ingressSGRulesInput := securityGroupRulesInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixIngress),
	}
	if err := s.deleteSecurityGroupRules(ingressSGRulesInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete rules for security group '%s': '%#v'", mastersSGRulesInput.GroupName, err))
	}

	// Delete masters security group.
	mastersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixMaster),
	}
	if err := s.deleteSecurityGroup(mastersSGInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': '%#v'", mastersSGInput.GroupName, err))
	}

	// Delete workers security group.
	workersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixWorker),
	}
	if err := s.deleteSecurityGroup(workersSGInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': '%#v'", workersSGInput.GroupName, err))
	}

	// Delete ingress security group.
	ingressSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: key.SecurityGroupName(cluster, prefixIngress),
	}
	if err := s.deleteSecurityGroup(ingressSGInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': '%#v'", ingressSGInput.GroupName, err))
	}

	vpcPeeringConection := &awsresources.VPCPeeringConnection{
		VPCId:     vpcID,
		PeerVPCId: cluster.Spec.AWS.VPC.PeerID,
		AWSEntity: awsresources.AWSEntity{
			Clients:     clients,
			HostClients: hostClients,
		},
	}
	conn, err := vpcPeeringConection.FindExisting()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not find vpc peering connection: '%#v'", err))
	}

	// Delete Guest VPC Routes.
	for _, privateRouteTableName := range cluster.Spec.AWS.VPC.RouteTableNames {
		privateRouteTable := &awsresources.RouteTable{
			Name:   privateRouteTableName,
			VpcID:  cluster.Spec.AWS.VPC.PeerID,
			Client: hostClients.EC2,
		}

		privateRoute := &awsresources.Route{
			RouteTable:           *privateRouteTable,
			DestinationCidrBlock: *conn.RequesterVpcInfo.CidrBlock,
			VpcID:                *conn.VpcPeeringConnectionId,
			AWSEntity:            awsresources.AWSEntity{Clients: hostClients},
		}

		if err := privateRoute.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete vpc route: '%v'", err))
		}
	}

	// Delete VPC peering connection.
	if err := vpcPeeringConection.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete vpc peering connection: '%#v'", err))
	} else {
		s.logger.Log("info", "deleted vpc peering connection")
	}

	// Delete VPC.
	if err := vpc.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete vpc: '%#v'", err))
	} else {
		s.logger.Log("info", "deleted vpc")
	}

	// Delete S3 bucket.
	bucketName := s.bucketName(cluster)

	bucket := &awsresources.Bucket{
		AWSEntity: awsresources.AWSEntity{Clients: clients},
		Name:      bucketName,
	}

	if err := bucket.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	}

	s.logger.Log("info", "deleted bucket")

	// Delete master policy.
	var masterPolicy resources.NamedResource
	masterPolicy = &awsresources.Policy{
		ClusterID:  key.ClusterID(cluster),
		PolicyType: prefixMaster,
		S3Bucket:   bucketName,
		AWSEntity:  awsresources.AWSEntity{Clients: clients},
	}
	if err := masterPolicy.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", fmt.Sprintf("deleted %s roles, policies, instance profiles", prefixMaster))
	}

	// Delete worker policy.
	var workerPolicy resources.NamedResource
	workerPolicy = &awsresources.Policy{
		ClusterID:  key.ClusterID(cluster),
		PolicyType: prefixWorker,
		S3Bucket:   bucketName,
		AWSEntity:  awsresources.AWSEntity{Clients: clients},
	}
	if err := workerPolicy.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", fmt.Sprintf("deleted %s roles, policies, instance profiles", prefixWorker))
	}

	// Delete KMS key.
	var kmsKey resources.ArnResource
	kmsKey = &awsresources.KMSKey{
		Name:      key.ClusterID(cluster),
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	if err := kmsKey.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", "deleted KMS key")
	}

	// Delete keypair.
	var keyPair resources.Resource
	keyPair = &awsresources.KeyPair{
		ClusterName: key.ClusterID(cluster),
		AWSEntity:   awsresources.AWSEntity{Clients: clients},
	}
	if err := keyPair.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", "deleted keypair")
	}

	s.logger.Log("info", fmt.Sprintf("cluster '%s' deleted", key.ClusterID(cluster)))
}

// TODO we need to support this in operatorkit.
func (s *Service) updateFunc(oldObj, newObj interface{}) {
	oldCluster := *oldObj.(*awstpr.CustomObject)
	cluster := *newObj.(*awstpr.CustomObject)

	if err := validateCluster(cluster); err != nil {
		s.logger.Log("error", "cluster spec is invalid: '%#v'", err)
		return
	}

	oldSize := key.WorkerCount(oldCluster)
	newSize := key.WorkerCount(cluster)

	if oldSize == newSize {
		// We get update events for all sorts of changes. We are currently only
		// interested in changes to one property, so we ignore all the others.
		return
	}

	s.awsConfig.Region = cluster.Spec.AWS.Region
	clients := awsutil.NewClients(s.awsConfig)

	err := s.awsConfig.SetAccountID(clients.IAM)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve amazon account id: '%#v'", err))
		return
	}

	asg := awsresources.AutoScalingGroup{
		Client:  clients.AutoScaling,
		Name:    fmt.Sprintf("%s-%s", key.ClusterID(cluster), prefixWorker),
		MinSize: newSize,
		MaxSize: newSize,
	}

	if err := asg.Update(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
		return
	}

	s.logger.Log("info", fmt.Sprintf("scaling workers auto scaling group from %d to %d", oldSize, newSize))
}
