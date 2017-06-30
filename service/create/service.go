package create

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/awstpr"
	awsinfo "github.com/giantswarm/awstpr/aws"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/clustertpr/node"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	certkit "github.com/giantswarm/operatorkit/secret/cert"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/juju/errgo"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

const (
	ClusterListAPIEndpoint  string = "/apis/cluster.giantswarm.io/v1/awses"
	ClusterWatchAPIEndpoint string = "/apis/cluster.giantswarm.io/v1/watch/awses"
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
	CertWatcher *certkit.Service
	K8sClient   kubernetes.Interface
	Logger      micrologger.Logger

	// Settings.
	AwsConfig  awsutil.Config
	PubKeyFile string
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		K8sClient:   nil,
		Logger:      nil,

		// Settings.
		AwsConfig:  awsutil.Config{},
		PubKeyFile: "",
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.AwsConfig must not be empty")
	}
	if config.PubKeyFile == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.PubKeyFile must not be empty")
	}

	newService := &Service{
		// Dependencies.
		certWatcher: config.CertWatcher,
		k8sClient:   config.K8sClient,
		logger:      config.Logger,

		// Internals
		bootOnce: sync.Once{},

		// Settings.
		awsConfig:  config.AwsConfig,
		pubKeyFile: config.PubKeyFile,
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	certWatcher *certkit.Service
	k8sClient   kubernetes.Interface
	logger      micrologger.Logger

	// Internals.
	bootOnce sync.Once

	// Settings.
	awsConfig  awsutil.Config
	pubKeyFile string
}

type Event struct {
	Type   string
	Object *awstpr.CustomObject
}

func (s *Service) newClusterListWatch() *cache.ListWatch {
	client := s.k8sClient.Core().RESTClient()

	listWatch := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			req := client.Get().AbsPath(ClusterListAPIEndpoint)
			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}

			var c awstpr.List
			if err := json.Unmarshal(b, &c); err != nil {
				return nil, err
			}

			return &c, nil
		},

		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			req := client.Get().AbsPath(ClusterWatchAPIEndpoint)
			return req.Watch()
		},
	}

	return listWatch
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		if err := s.createTPR(); err != nil {
			panic(err)
		}
		s.logger.Log("info", "successfully created third-party resource")

		_, clusterInformer := cache.NewInformer(
			s.newClusterListWatch(),
			&awstpr.CustomObject{},
			tpr.ResyncPeriod,
			cache.ResourceEventHandlerFuncs{
				AddFunc:    s.onAdd,
				UpdateFunc: s.onUpdate,
				DeleteFunc: s.onDelete,
			},
		)

		s.logger.Log("info", "starting watch")

		// Cluster informer lifecycle can be interrupted by putting a value into a "stop channel".
		// We aren't currently using that functionality, so we are passing a nil here.
		clusterInformer.Run(nil)
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

		machines    []node.Node
		awsMachines []awsinfo.Node
		instanceIDs []string

		extension cloudconfig.Extension
	)

	switch input.prefix {
	case prefixMaster:
		machines = input.cluster.Spec.Cluster.Masters
		awsMachines = input.cluster.Spec.AWS.Masters
		extension = NewMasterCloudConfigExtension(input.cluster.Spec, input.tlsAssets)
	case prefixWorker:
		machines = input.cluster.Spec.Cluster.Workers
		awsMachines = input.cluster.Spec.AWS.Workers
		extension = NewWorkerCloudConfigExtension(input.cluster.Spec, input.tlsAssets)
	}

	// TODO(nhlfr): Create a separate module for validating specs and execute on the earlier stages.
	if len(machines) != len(awsMachines) {
		return false, nil, microerror.MaskAny(fmt.Errorf("mismatched number of %s machines in the 'spec' and 'aws' sections: %d != %d",
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
			extension:           extension,
			tlsAssets:           input.tlsAssets,
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
			return false, nil, microerror.MaskAny(err)
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
	cluster             awstpr.CustomObject
	machine             node.Node
	awsNode             awsinfo.Node
	extension           cloudconfig.Extension
	tlsAssets           *certificatetpr.CompactTLSAssets
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
	cloudConfigParams := cloudconfig.Params{
		Cluster:   input.cluster.Spec.Cluster,
		Node:      input.machine,
		Extension: input.extension,
	}

	cloudConfig, err := s.cloudConfig(input.prefix, cloudConfigParams, input.cluster.Spec, input.tlsAssets)
	if err != nil {
		return false, "", microerror.MaskAny(err)
	}

	// We now upload the instance cloudconfig to S3 and create a "small
	// cloudconfig" that just fetches the previously uploaded "final
	// cloudconfig" and executes coreos-cloudinit with it as argument.
	// We do this to circumvent the 16KB limit on user-data for EC2 instances.

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
		return false, "", microerror.MaskAny(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		return false, "", microerror.MaskAny(err)
	}

	securityGroupID, err := input.securityGroup.GetID()
	if err != nil {
		return false, "", microerror.MaskAny(err)
	}

	subnetID, err := input.subnet.GetID()
	if err != nil {
		return false, "", microerror.MaskAny(err)
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
			return false, "", microerror.MaskAny(err)
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
		return microerror.MaskAny(err)
	}

	for _, instance := range instances {
		if err := instance.Delete(); err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

type deleteMachineInput struct {
	name    string
	clients awsutil.Clients
	machine node.Node
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

func (s *Service) onAdd(obj interface{}) {
	cluster := *obj.(*awstpr.CustomObject)
	s.logger.Log("info", fmt.Sprintf("creating cluster '%s'", cluster.Name))

	if err := validateCluster(cluster); err != nil {
		s.logger.Log("error", "cluster spec is invalid: %s", errgo.Details(err))
		return
	}

	if err := s.createClusterNamespace(cluster.Spec.Cluster); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create cluster namespace: %s", errgo.Details(err)))
		return
	}

	// Create AWS client
	s.awsConfig.Region = cluster.Spec.AWS.Region
	clients := awsutil.NewClients(s.awsConfig)

	err := s.awsConfig.SetAccountID(clients.IAM)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve amazon account id: %s", errgo.Details(err)))
		return
	}

	// Create keypair
	var keyPair resources.ReusableResource
	var keyPairCreated bool
	{
		var err error
		keyPair = &awsresources.KeyPair{
			ClusterName: cluster.Name,
			Provider:    awsresources.NewFSKeyPairProvider(s.pubKeyFile),
			AWSEntity:   awsresources.AWSEntity{Clients: clients},
		}
		keyPairCreated, err = keyPair.CreateIfNotExists()
		if err != nil {
			s.logger.Log("error", fmt.Sprintf("could not create keypair: %s", errgo.Details(err)))
			return
		}
	}

	if keyPairCreated {
		s.logger.Log("info", fmt.Sprintf("created keypair '%s'", cluster.Name))
	} else {
		s.logger.Log("info", fmt.Sprintf("keypair '%s' already exists, reusing", cluster.Name))
	}

	s.logger.Log("info", fmt.Sprintf("waiting for k8s secrets..."))
	clusterID := cluster.Spec.Cluster.Cluster.ID
	certs, err := s.certWatcher.SearchCerts(clusterID)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not get certificates from secrets: %v", errgo.Details(err)))
		return
	}

	// Create KMS key
	kmsKey := &awsresources.KMSKey{
		Name:      cluster.Name,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}

	kmsCreated, kmsKeyErr := kmsKey.CreateIfNotExists()
	if kmsKeyErr != nil {
		s.logger.Log("error", fmt.Sprintf("could not create KMS key: %v", errgo.Details(kmsKeyErr)))
		return
	}

	if kmsCreated {
		s.logger.Log("info", fmt.Sprintf("created KMS key for cluster '%s'", cluster.Name))
	} else {
		s.logger.Log("info", fmt.Sprintf("kms key '%s' already exists, reusing", kmsKey.Name))
	}

	// Encode TLS assets
	tlsAssets, err := s.encodeTLSAssets(certs, clients.KMS, kmsKey.Arn())
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not encode TLS assets: %s", errgo.Details(err)))
		return
	}

	// Create policy
	bucketName := s.bucketName(cluster)

	var policy resources.NamedResource
	var policyErr error
	{
		policy = &awsresources.Policy{
			ClusterID: cluster.Spec.Cluster.Cluster.ID,
			KMSKeyArn: kmsKey.Arn(),
			S3Bucket:  bucketName,
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}
		policyErr = policy.CreateOrFail()
	}
	if policyErr != nil {
		s.logger.Log("error", fmt.Sprintf("could not create policy: %s", errgo.Details(policyErr)))
	}

	// Create S3 bucket
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
			s.logger.Log("error", fmt.Sprintf("could not create S3 bucket: %s", errgo.Details(err)))
			return
		}
	}

	if bucketCreated {
		s.logger.Log("info", fmt.Sprintf("created bucket '%s'", bucketName))
	} else {
		s.logger.Log("info", fmt.Sprintf("bucket '%s' already exists, reusing", bucketName))
	}

	// Create VPC
	var vpc resources.ResourceWithID
	vpc = &awsresources.VPC{
		CidrBlock: cluster.Spec.AWS.VPC.CIDR,
		Name:      cluster.Name,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	vpcCreated, err := vpc.CreateIfNotExists()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create VPC: %s", errgo.Details(err)))
		return
	}
	if vpcCreated {
		s.logger.Log("info", fmt.Sprintf("created vpc for cluster '%s'", cluster.Name))
	} else {
		s.logger.Log("info", fmt.Sprintf("vpc for cluster '%s' already exists, reusing", cluster.Name))
	}
	vpcID, err := vpc.GetID()
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
	}

	// Create gateway
	var gateway resources.ResourceWithID
	gateway = &awsresources.Gateway{
		Name:  cluster.Name,
		VpcID: vpcID,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	gatewayCreated, err := gateway.CreateIfNotExists()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create gateway: %s", errgo.Details(err)))
		return
	}
	if gatewayCreated {
		s.logger.Log("info", fmt.Sprintf("created gateway for cluster '%s'", cluster.Name))
	} else {
		s.logger.Log("info", fmt.Sprintf("gateway for cluster '%s' already exists, reusing", cluster.Name))
	}

	// Create masters security group.
	mastersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: securityGroupName(cluster.Name, prefixMaster),
		VPCID:     vpcID,
	}
	mastersSecurityGroup, err := s.createSecurityGroup(mastersSGInput)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create security group '%s': %s", mastersSGInput.GroupName, errgo.Details(err)))
		return
	}
	mastersSecurityGroupID, err := mastersSecurityGroup.GetID()
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	// Create workers security group.
	workersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: securityGroupName(cluster.Name, prefixWorker),
		VPCID:     vpcID,
	}
	workersSecurityGroup, err := s.createSecurityGroup(workersSGInput)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create security group '%s': %s", workersSGInput.GroupName, errgo.Details(err)))
		return
	}
	workersSecurityGroupID, err := workersSecurityGroup.GetID()
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	// Create ingress ELB security group.
	ingressSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: securityGroupName(cluster.Name, prefixIngress),
		VPCID:     vpcID,
	}
	ingressSecurityGroup, err := s.createSecurityGroup(ingressSGInput)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create security group '%s': %s", ingressSGInput.GroupName, errgo.Details(err)))
		return
	}
	ingressSecurityGroupID, err := ingressSecurityGroup.GetID()
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	// Create rules for the security groups.
	rulesInput := rulesInput{
		Cluster:                cluster,
		MastersSecurityGroupID: mastersSecurityGroupID,
		WorkersSecurityGroupID: workersSecurityGroupID,
		IngressSecurityGroupID: ingressSecurityGroupID,
	}

	if err := mastersSecurityGroup.ApplyRules(rulesInput.masterRules()); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create rules for security group '%s': %s", mastersSecurityGroup.GroupName, errgo.Details(err)))
		return
	}

	if err := workersSecurityGroup.ApplyRules(rulesInput.workerRules()); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create rules for security group '%s': %s", workersSecurityGroup.GroupName, errgo.Details(err)))
		return
	}

	if err := ingressSecurityGroup.ApplyRules(rulesInput.ingressRules()); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create rules for security group '%s': %s", ingressSecurityGroup.GroupName, errgo.Details(err)))
		return
	}

	// Create route table.
	routeTable := &awsresources.RouteTable{
		Name:   cluster.Name,
		VpcID:  vpcID,
		Client: clients.EC2,
	}
	routeTableCreated, err := routeTable.CreateIfNotExists()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create route table: %s", errgo.Details(err)))
		return
	}
	if routeTableCreated {
		s.logger.Log("info", "created route table")
	} else {
		s.logger.Log("info", "route table already exists, reusing")
	}

	if err := routeTable.MakePublic(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not make route table public: %s", errgo.Details(err)))
		return
	}

	// Create public subnet for the masters
	publicSubnet := &awsresources.Subnet{
		AvailabilityZone: cluster.Spec.AWS.AZ,
		CidrBlock:        cluster.Spec.AWS.VPC.PublicSubnetCIDR,
		Name:             subnetName(cluster, suffixPublic),
		VpcID:            vpcID,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	publicSubnetCreated, err := publicSubnet.CreateIfNotExists()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create public subnet: %s", errgo.Details(err)))
		return
	}
	if publicSubnetCreated {
		s.logger.Log("info", fmt.Sprintf("created public subnet for cluster '%s'", cluster.Name))
	} else {
		s.logger.Log("info", fmt.Sprintf("public subnet for cluster '%s' already exists, reusing", cluster.Name))
	}
	publicSubnetID, err := publicSubnet.GetID()
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	if err := publicSubnet.MakePublic(routeTable); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not make subnet public, %s", errgo.Details(err)))
		return
	}

	// Run masters
	anyMastersCreated, masterIDs, err := s.runMachines(runMachinesInput{
		clients:             clients,
		cluster:             cluster,
		tlsAssets:           tlsAssets,
		clusterName:         cluster.Name,
		bucket:              bucket,
		securityGroup:       mastersSecurityGroup,
		subnet:              publicSubnet,
		keyPairName:         cluster.Name,
		instanceProfileName: policy.GetName(),
		prefix:              prefixMaster,
	})
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
	}

	if !validateIDs(masterIDs) {
		s.logger.Log("error", fmt.Sprintf("master nodes had invalid instance IDs: %v", masterIDs))
		return
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
		s.logger.Log("error", errgo.Details(err))
		return
	}

	// Assign the ProxyProtocol policy to the apiserver load balancer.
	if err := apiLB.AssignProxyProtocolPolicy(); err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
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

	etcdLB, err := s.createLoadBalancer(lbInput)
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	// If the policy couldn't be created and some instances didn't exist before, that means that the cluster
	// is inconsistent and most problably its deployment broke in the middle during the previous run of
	// aws-operator.
	if anyMastersCreated && (kmsKeyErr != nil || policyErr != nil) {
		s.logger.Log("error", fmt.Sprintf("cluster '%s' is inconsistent, KMS keys and policies were not created, but EC2 instances were missing, please consider deleting this cluster", cluster.Name))
		return
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
		s.logger.Log("error", errgo.Details(err))
		return
	}

	// Assign the ProxyProtocol policy to the Ingress load balancer.
	if err := ingressLB.AssignProxyProtocolPolicy(); err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	s.logger.Log("info", fmt.Sprintf("created ingress load balancer"))

	s.logger.Log("info", fmt.Sprintf("creating workers auto scaling group..."))
	// Create an Auto Scaling Group for the workers.
	asgStackInput := asgStackInput{
		// Dependencies.
		clients: clients,

		// Settings.
		bucket:                 bucket,
		cluster:                cluster,
		iamInstanceProfileName: policy.GetName(),
		keyPairName:            cluster.Name,
		prefix:                 prefixWorker,
		workersSecurityGroupRules: rulesInput.workerRules(),
		ingressSecurityGroupRules: rulesInput.ingressRules(),
		subnetID:                  publicSubnetID,
		tlsAssets:                 tlsAssets,
		vpcID:                     vpcID,
		elbListeners: []awsresources.PortPair{
			{
				PortELB:      httpsPort,
				PortInstance: cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
			},
			{
				PortELB:      httpPort,
				PortInstance: cluster.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
			},
		},
	}

	if err := s.createASGStack(asgStackInput); err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	s.logger.Log("info", fmt.Sprintf("created workers auto scaling group"))

	// Create Record Sets for the Load Balancers.
	recordSetInputs := []recordSetInput{
		recordSetInput{
			Cluster:      cluster,
			Client:       clients.Route53,
			Resource:     apiLB,
			Domain:       cluster.Spec.Cluster.Kubernetes.API.Domain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.API,
			Type:         route53.RRTypeA,
		},
		recordSetInput{
			Cluster:      cluster,
			Client:       clients.Route53,
			Resource:     etcdLB,
			Domain:       cluster.Spec.Cluster.Etcd.Domain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.Etcd,
			Type:         route53.RRTypeA,
		},
		recordSetInput{
			Cluster:      cluster,
			Client:       clients.Route53,
			Resource:     ingressLB,
			Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
			HostedZoneID: cluster.Spec.AWS.HostedZones.Ingress,
			Type:         route53.RRTypeA,
		},
		recordSetInput{
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
			s.logger.Log("error", errgo.Details(rsErr))
			return
		}
	}
	if rsErr == nil {
		s.logger.Log("info", fmt.Sprintf("created DNS records for load balancers"))
	}

	s.logger.Log("info", fmt.Sprintf("cluster '%s' processed", cluster.Name))
}

func (s *Service) onDelete(obj interface{}) {
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
		s.logger.Log("error", "cluster spec is invalid: %s", errgo.Details(err))
		return
	}

	if err := s.deleteClusterNamespace(cluster.Spec.Cluster); err != nil {
		s.logger.Log("error", "could not delete cluster namespace:", err)
	}

	clients := awsutil.NewClients(s.awsConfig)

	err := s.awsConfig.SetAccountID(clients.IAM)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve amazon account id: %s", errgo.Details(err)))
		return
	}

	// Delete masters.
	s.logger.Log("info", "deleting masters...")
	if err := s.deleteMachines(deleteMachinesInput{
		clients:     clients,
		clusterName: cluster.Name,
		prefix:      prefixMaster,
	}); err != nil {
		s.logger.Log("error", errgo.Details(err))
	} else {
		s.logger.Log("info", "deleted masters")
	}

	// Delete workers Auto Scaling Group stack.
	stack := awsresources.ASGStack{
		Client: clients.CloudFormation,
		Name:   fmt.Sprintf("%s-%s", cluster.Name, prefixWorker),
	}

	if err := stack.Delete(); err != nil {
		s.logger.Log("error", errgo.Details(err))
	} else {
		s.logger.Log("info", "deleted workers auto scaling group")
	}

	// Delete Record Sets.
	apiLBName, err := loadBalancerName(cluster.Spec.Cluster.Kubernetes.API.Domain, cluster)
	etcdLBName, err := loadBalancerName(cluster.Spec.Cluster.Etcd.Domain, cluster)
	ingressLBName, err := loadBalancerName(cluster.Spec.Cluster.Kubernetes.IngressController.Domain, cluster)
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
	} else {
		apiLB, err := awsresources.NewELBFromExisting(apiLBName, clients.ELB)
		etcdLB, err := awsresources.NewELBFromExisting(etcdLBName, clients.ELB)
		ingressLB, err := awsresources.NewELBFromExisting(ingressLBName, clients.ELB)
		if err != nil {
			s.logger.Log("error", errgo.Details(err))
		} else {
			recordSetInputs := []recordSetInput{
				recordSetInput{
					Cluster:      cluster,
					Client:       clients.Route53,
					Resource:     apiLB,
					Domain:       cluster.Spec.Cluster.Kubernetes.API.Domain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.API,
				},
				recordSetInput{
					Cluster:      cluster,
					Client:       clients.Route53,
					Resource:     etcdLB,
					Domain:       cluster.Spec.Cluster.Etcd.Domain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.Etcd,
				},
				recordSetInput{
					Cluster:      cluster,
					Client:       clients.Route53,
					Resource:     ingressLB,
					Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
					HostedZoneID: cluster.Spec.AWS.HostedZones.Ingress,
				},
				recordSetInput{
					Cluster: cluster,
					Client:  clients.Route53,
					Value:   cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
					Domain:  cluster.Spec.Cluster.Kubernetes.IngressController.WildcardDomain,
				},
			}

			var rsErr error
			for _, input := range recordSetInputs {
				if rsErr = s.deleteRecordSet(input); rsErr != nil {
					s.logger.Log("error", errgo.Details(rsErr))
				}
			}
			if rsErr == nil {
				s.logger.Log("info", "deleted record sets")
			}
		}
	}

	// Delete Load Balancers.
	loadBalancerInputs := []LoadBalancerInput{
		LoadBalancerInput{
			Name:    cluster.Spec.Cluster.Kubernetes.API.Domain,
			Clients: clients,
			Cluster: cluster,
		},
		LoadBalancerInput{
			Name:    cluster.Spec.Cluster.Etcd.Domain,
			Clients: clients,
			Cluster: cluster,
		},
		LoadBalancerInput{
			Name:    cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
			Clients: clients,
			Cluster: cluster,
		},
	}

	var elbErr error
	for _, lbInput := range loadBalancerInputs {
		if elbErr = s.deleteLoadBalancer(lbInput); elbErr != nil {
			s.logger.Log("error", errgo.Details(elbErr))
		}
	}
	if elbErr == nil {
		s.logger.Log("info", "deleted ELBs")
	}

	// Delete route table.
	var routeTable resources.ResourceWithID
	routeTable = &awsresources.RouteTable{
		Name:   cluster.Name,
		Client: clients.EC2,
	}
	if err := routeTable.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete route table: %s", errgo.Details(err)))
	} else {
		s.logger.Log("info", "deleted route table")
	}

	// Sync VPC
	var vpc resources.ResourceWithID
	vpc = &awsresources.VPC{
		Name:      cluster.Name,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	vpcID, err := vpc.GetID()
	if err != nil {
		s.logger.Log("error", errgo.Details(err))
	}

	// Delete gateway.
	var gateway resources.ResourceWithID
	gateway = &awsresources.Gateway{
		Name:  cluster.Name,
		VpcID: vpcID,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	if err := gateway.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete gateway: %s", errgo.Details(err)))
	} else {
		s.logger.Log("info", "deleted gateway")
	}

	// Delete public subnet.
	publicSubnet := &awsresources.Subnet{
		Name: subnetName(cluster, suffixPublic),
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	if err := publicSubnet.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete public subnet: %s", errgo.Details(err)))
	} else {
		s.logger.Log("info", "deleted public subnet")
	}

	// Delete masters security group.
	mastersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: securityGroupName(cluster.Name, prefixMaster),
	}
	if err := s.deleteSecurityGroup(mastersSGInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': %s", mastersSGInput.GroupName, errgo.Details(err)))
	}

	// Delete workers security group.
	workersSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: securityGroupName(cluster.Name, prefixWorker),
	}
	if err := s.deleteSecurityGroup(workersSGInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': %s", workersSGInput.GroupName, errgo.Details(err)))
	}

	// Delete ingress security group.
	ingressSGInput := securityGroupInput{
		Clients:   clients,
		GroupName: securityGroupName(cluster.Name, prefixIngress),
	}
	if err := s.deleteSecurityGroup(ingressSGInput); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': %s", ingressSGInput.GroupName, errgo.Details(err)))
	}

	// Delete VPC.
	if err := vpc.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete vpc: %s", errgo.Details(err)))
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
		s.logger.Log("error", errgo.Details(err))
	}

	s.logger.Log("info", "deleted bucket")

	// Delete policy.
	var policy resources.NamedResource
	policy = &awsresources.Policy{
		ClusterID: cluster.Spec.Cluster.Cluster.ID,
		S3Bucket:  bucketName,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	if err := policy.Delete(); err != nil {
		s.logger.Log("error", errgo.Details(err))
	} else {
		s.logger.Log("info", "deleted roles, policies, instance profiles")
	}

	// Delete KMS key.
	var kmsKey resources.ArnResource
	kmsKey = &awsresources.KMSKey{
		Name:      cluster.Name,
		AWSEntity: awsresources.AWSEntity{Clients: clients},
	}
	if err := kmsKey.Delete(); err != nil {
		s.logger.Log("error", errgo.Details(err))
	} else {
		s.logger.Log("info", "deleted KMS key")
	}

	// Delete keypair.
	var keyPair resources.Resource
	keyPair = &awsresources.KeyPair{
		ClusterName: cluster.Name,
		AWSEntity:   awsresources.AWSEntity{Clients: clients},
	}
	if err := keyPair.Delete(); err != nil {
		s.logger.Log("error", errgo.Details(err))
	} else {
		s.logger.Log("info", "deleted keypair")
	}

	s.logger.Log("info", fmt.Sprintf("cluster '%s' deleted", cluster.Name))
}

func (s *Service) onUpdate(oldObj, newObj interface{}) {
	oldCluster := *oldObj.(*awstpr.CustomObject)
	cluster := *newObj.(*awstpr.CustomObject)

	if err := validateCluster(cluster); err != nil {
		s.logger.Log("error", "cluster spec is invalid: %s", errgo.Details(err))
		return
	}

	oldSize := len(oldCluster.Spec.AWS.Workers)
	newSize := len(cluster.Spec.AWS.Workers)

	if oldSize == newSize {
		// We get update events for all sorts of changes. We are currently only
		// interested in changes to one property, so we ignore all the others.
		return
	} else if oldSize > newSize {
		s.logger.Log("clusterID", cluster.Spec.Cluster.Cluster.ID, "info", "cluster size reduction not supported")
		return
	}

	s.awsConfig.Region = cluster.Spec.AWS.Region
	clients := awsutil.NewClients(s.awsConfig)

	err := s.awsConfig.SetAccountID(clients.IAM)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve amazon account id: %s", errgo.Details(err)))
		return
	}

	asg := awsresources.AutoScalingGroup{
		Client:  clients.AutoScaling,
		Name:    s.asgName(cluster, prefixWorker),
		MinSize: newSize,
		MaxSize: newSize,
	}

	if err := asg.Update(); err != nil {
		s.logger.Log("error", errgo.Details(err))
		return
	}

	s.logger.Log("info", fmt.Sprintf("increased size of workers auto scaling group to %d", newSize))
}
