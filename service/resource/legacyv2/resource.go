package legacyv2

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfigv2"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/aws-operator/service/resource/legacyv2/adapter"
)

const (
	// Name is the identifier of the resource.
	Name = "legacyv2"

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

// Config represents the configuration used to create a legacy service.
type Config struct {
	// Dependencies.
	CertWatcher *certificatetpr.Service
	CloudConfig *cloudconfigv2.CloudConfig
	K8sClient   kubernetes.Interface
	KeyWatcher  *randomkeytpr.Service
	Logger      micrologger.Logger
	Clients     *adapter.Clients
	HostClients *adapter.Clients

	// Settings.
	AwsConfig        awsutil.Config
	AwsHostConfig    awsutil.Config
	InstallationName string
	PubKeyFile       string
}

// DefaultConfig provides a default configuration to create a new legacy
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		CloudConfig: nil,
		K8sClient:   nil,
		KeyWatcher:  nil,
		Logger:      nil,
		Clients:     nil,
		HostClients: nil,

		// Settings.
		AwsConfig:        awsutil.Config{},
		AwsHostConfig:    awsutil.Config{},
		InstallationName: "",
		PubKeyFile:       "",
	}
}

// New creates a new configured resource.
func New(config Config) (*Resource, error) {
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
	if config.Clients == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Clients must not be empty")
	}
	if config.HostClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.HostClients must not be empty")
	}

	// Settings.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsConfig must not be empty")
	}
	if config.AwsHostConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsHostConfig must not be empty")
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.InstallationName must not be empty")
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

	newService := &Resource{
		// Dependencies.
		certWatcher:    config.CertWatcher,
		cloudConfig:    config.CloudConfig,
		k8sClient:      config.K8sClient,
		keyWatcher:     config.KeyWatcher,
		logger:         config.Logger,
		awsClients:     config.Clients,
		awsHostClients: config.HostClients,

		// Internals
		bootOnce: sync.Once{},
		tpr:      newTPR,

		// Settings.
		awsConfig:        config.AwsConfig,
		awsHostConfig:    config.AwsHostConfig,
		installationName: config.InstallationName,
		pubKeyFile:       config.PubKeyFile,
	}

	return newService, nil
}

// Resource implements the legacy resource.
type Resource struct {
	// Dependencies.
	certWatcher    *certificatetpr.Service
	cloudConfig    *cloudconfigv2.CloudConfig
	k8sClient      kubernetes.Interface
	keyWatcher     *randomkeytpr.Service
	logger         micrologger.Logger
	awsClients     *adapter.Clients
	awsHostClients *adapter.Clients

	// Internals.
	bootOnce sync.Once
	tpr      *tpr.TPR

	// Settings.
	awsConfig        awsutil.Config
	awsHostConfig    awsutil.Config
	installationName string
	pubKeyFile       string
}

// NewUpdatePatch is called upon observed custom object change. It creates the
// AWS resources for the cluster.
func (s *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	customObject, ok := obj.(*v1alpha1.AWSConfig)
	if !ok {
		return &framework.Patch{}, microerror.Maskf(invalidConfigError, "could not convert to v1alpha1.AWSConfig")
	}
	cluster := *customObject

	s.logger.Log("info", fmt.Sprintf("updating cluster '%s'", keyv2.ClusterID(cluster)))

	// legacy logic
	if err := validateCluster(cluster); err != nil {
		return nil, microerror.Mask(err)
	}

	err := s.processCluster(cluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s.logger.Log("info", fmt.Sprintf("cluster '%s' processed", keyv2.ClusterID(cluster)))

	patch := framework.NewPatch()

	// cloudformation logic: on creation we need to create the cloudformation resource after legacy so that there are no dependency
	// problems (see https://github.com/giantswarm/operatorkit/issues/139). Once the transition to cloudformation is done we
	// will separate the cloudformation and legacy resources.
	if keyv2.UseCloudFormation(cluster) {
		create, err := s.newCreateChange(ctx, obj, currentState, desiredState)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		update, err := s.newUpdateChange(ctx, obj, currentState, desiredState)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		patch.SetCreateChange(create)
		patch.SetUpdateChange(update)
	}
	return patch, nil
}

// NewDeletePatch is called upon observed custom object deletion. It deletes the
// AWS resources for the cluster.
func (s *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	// We can receive an instance of v1alpha1.AWSConfig or cache.DeletedFinalStateUnknown.
	// We need to assert the type properly and log an error when we cannot do that.
	// Also, the cache.DeleteFinalStateUnknown object can contain the proper CustomObject,
	// but doesn't have to.
	// https://github.com/kubernetes/client-go/blob/7ba6be594966f4bec08a57a60c855d17a9f7000a/tools/cache/delta_fifo.go#L674-L677
	var cluster v1alpha1.AWSConfig
	clusterPtr, ok := obj.(*v1alpha1.AWSConfig)
	if ok {
		cluster = *clusterPtr
	} else {
		deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, microerror.Maskf(invalidConfigError, "received unknown type of third-party object")
		}
		clusterPtr, ok := deletedObj.Obj.(*v1alpha1.AWSConfig)
		if !ok {
			return nil, microerror.Maskf(invalidConfigError, "received the proper delete request, but the type of third-party object is unknown")
		}
		cluster = *clusterPtr
	}

	s.logger.Log("info", fmt.Sprintf("deleting cluster '%s'", keyv2.ClusterID(cluster)))

	// cloudformation logic: on deletion we need to remove first the cloudformation resource so that there are no dependency
	// problems (see https://github.com/giantswarm/operatorkit/issues/139). Once the transition to cloudformation is done we
	// will separate the cloudformation and legacy resources.
	if keyv2.UseCloudFormation(cluster) {
		deleteStackInput := cloudformation.DeleteStackInput{
			StackName: aws.String(keyv2.MainGuestStackName(cluster)),
		}
		_, err := s.awsClients.CloudFormation.DeleteStack(&deleteStackInput)
		if err != nil {
			return nil, microerror.Maskf(err, "deleting AWS Guest CloudFormation Stack")
		}
		s.logger.LogCtx(ctx, "debug", "deleting AWS Guest CloudFormation stack: deleted")

		deleteStackInput = cloudformation.DeleteStackInput{
			StackName: aws.String(keyv2.MainHostPreStackName(cluster)),
		}
		_, err = s.awsHostClients.CloudFormation.DeleteStack(&deleteStackInput)
		if err != nil {
			return nil, microerror.Maskf(err, "deleting AWS Host Pre-Guest CloudFormation Stack")
		}
		s.logger.LogCtx(ctx, "debug", "deleting AWS Host Pre-Guest CloudFormation stack: deleted")

		deleteStackInput = cloudformation.DeleteStackInput{
			StackName: aws.String(keyv2.MainHostPostStackName(cluster)),
		}
		_, err = s.awsHostClients.CloudFormation.DeleteStack(&deleteStackInput)
		if err != nil {
			return nil, microerror.Maskf(err, "deleting AWS Host Post-Guest CloudFormation Stack")
		}
		s.logger.LogCtx(ctx, "debug", "deleting AWS Host Post-Guest CloudFormation stack: deleted")
	}

	// legacy logic
	err := s.processDelete(cluster)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("error deleting cluster '%s': '%#v'", keyv2.ClusterID(cluster), err))
		return nil, microerror.Mask(err)
	}

	s.logger.Log("info", fmt.Sprintf("cluster '%s' deleted", keyv2.ClusterID(cluster)))

	return &framework.Patch{}, nil
}

func (s *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cluster, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	if keyv2.UseCloudFormation(cluster) {
		updateStackInput, err := toUpdateStackInput(updateChange)
		if err != nil {
			return microerror.Mask(err)
		}

		stackName := updateStackInput.StackName
		if *stackName != "" {
			_, err := s.awsClients.CloudFormation.UpdateStack(&updateStackInput)
			if err != nil {
				return microerror.Maskf(err, "updating AWS cloudformation stack")
			}

			s.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: updated")
		} else {
			s.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: no need to update")
		}
	}
	return nil
}

func (s *Resource) processCluster(cluster v1alpha1.AWSConfig) error {
	// For new clusters using Cloud Formation there is an OperatorKit resource
	// for the k8s namespace.
	if !keyv2.UseCloudFormation(cluster) {
		// Create cluster namespace in k8s.
		if err := s.createClusterNamespace(cluster.Spec.Cluster); err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create cluster namespace: '%#v'", err))
		}
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
	if !keyv2.HasClusterVersion(cluster) {
		// Create keypair.
		var keyPair resources.ReusableResource
		var keyPairCreated bool
		{
			var err error
			keyPair = &awsresources.KeyPair{
				ClusterName: keyv2.ClusterID(cluster),
				Provider:    awsresources.NewFSKeyPairProvider(s.pubKeyFile),
				AWSEntity:   awsresources.AWSEntity{Clients: clients},
			}
			keyPairCreated, err = keyPair.CreateIfNotExists()
			if err != nil {
				return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create keypair: '%#v'", err))
			}
		}

		if keyPairCreated {
			s.logger.Log("info", fmt.Sprintf("created keypair '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("keypair '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}
	}

	// For new clusters using Cloud Formation there is an OperatorKit resource
	// for the kms keys and related resources.
	var kmsKey *awsresources.KMSKey
	var tlsAssets *certificatetpr.CompactTLSAssets
	var clusterKeys *randomkeytpr.CompactRandomKeyAssets
	if !keyv2.UseCloudFormation(cluster) {
		s.logger.Log("info", fmt.Sprintf("waiting for k8s secrets..."))
		clusterID := keyv2.ClusterID(cluster)
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
		kmsKey = &awsresources.KMSKey{
			Name:      keyv2.ClusterID(cluster),
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}

		kmsCreated, kmsKeyErr := kmsKey.CreateIfNotExists()
		if kmsKeyErr != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create KMS key: '%#v'", kmsKeyErr))
		}

		if kmsCreated {
			s.logger.Log("info", fmt.Sprintf("created KMS key for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("kms key '%s' already exists, reusing", kmsKey.Name))
		}

		// Encode TLS assets.
		tlsAssets, err = s.encodeTLSAssets(certs, clients.KMS, kmsKey.Arn())
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not encode TLS assets: '%#v'", err))
		}

		// Encode Key assets.
		clusterKeys, err = s.encodeKeyAssets(keys, clients.KMS, kmsKey.Arn())
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not encode Keys assets: '%#v'", err))
		}
	}

	bucketName := s.bucketName(cluster)

	// Create master IAM policy.
	var masterPolicy resources.NamedResource
	var masterPolicyCreated bool
	var workerPolicy resources.NamedResource

	if !keyv2.UseCloudFormation(cluster) {
		{
			var err error
			masterPolicy = &awsresources.Policy{
				ClusterID:  keyv2.ClusterID(cluster),
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
			s.logger.Log("info", fmt.Sprintf("created master policy for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("master policy for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}

		// Create worker IAM policy.
		var workerPolicyCreated bool
		{
			var err error
			workerPolicy = &awsresources.Policy{
				ClusterID:  keyv2.ClusterID(cluster),
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
			s.logger.Log("info", fmt.Sprintf("created worker policy for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("worker policy for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}
	}

	// Create S3 bucket.
	var bucket resources.ReusableResource
	if !keyv2.UseCloudFormation(cluster) {
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
	}

	var vpcID string
	var conn *ec2.VpcPeeringConnection
	if !keyv2.UseCloudFormation(cluster) {
		// Create VPC.
		var vpc resources.ResourceWithID
		vpc = &awsresources.VPC{
			CidrBlock:        cluster.Spec.AWS.VPC.CIDR,
			InstallationName: s.installationName,
			Name:             keyv2.ClusterID(cluster),
			AWSEntity:        awsresources.AWSEntity{Clients: clients},
		}
		vpcCreated, err := vpc.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create VPC: '%#v'", err))
		}
		if vpcCreated {
			s.logger.Log("info", fmt.Sprintf("created vpc for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("vpc for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}
		vpcID, err = vpc.GetID()
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
			Logger: s.logger,
		}
		vpcPeeringConnectionCreated, err := vpcPeeringConection.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create vpc peering connection: '%#v'", err))
		}
		if vpcPeeringConnectionCreated {
			s.logger.Log("info", fmt.Sprintf("created vpc peering connection for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("vpc peering connection for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}

		conn, err = vpcPeeringConection.FindExisting()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not find vpc peering connection: '%#v'", err))
		}
	}

	if !keyv2.UseCloudFormation(cluster) {
		// Create internet gateway.
		var internetGateway resources.ResourceWithID
		internetGateway = &awsresources.InternetGateway{
			Name:  keyv2.ClusterID(cluster),
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
			s.logger.Log("info", fmt.Sprintf("created internet gateway for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("internet gateway for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}
	}

	var mastersSecurityGroup *awsresources.SecurityGroup
	var workersSecurityGroup *awsresources.SecurityGroup
	var ingressSecurityGroup *awsresources.SecurityGroup
	var mastersSecurityGroupID string
	var workersSecurityGroupID string
	var ingressSecurityGroupID string
	var err error
	if !keyv2.UseCloudFormation(cluster) {
		// Create masters security group.
		mastersSGInput := securityGroupInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixMaster),
			VPCID:     vpcID,
		}
		mastersSecurityGroup, err = s.createSecurityGroup(mastersSGInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create security group '%s': '%#v'", mastersSGInput.GroupName, err))
		}
		mastersSecurityGroupID, err = mastersSecurityGroup.GetID()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get security group '%s' ID: '%#v'", mastersSGInput.GroupName, err))
		}

		// Create workers security group.
		workersSGInput := securityGroupInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixWorker),
			VPCID:     vpcID,
		}
		workersSecurityGroup, err = s.createSecurityGroup(workersSGInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create security group '%s': '%#v'", workersSGInput.GroupName, err))
		}
		workersSecurityGroupID, err = workersSecurityGroup.GetID()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get security group '%s' ID: '%#v'", workersSGInput.GroupName, err))
		}

		// Create ingress ELB security group.
		ingressSGInput := securityGroupInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixIngress),
			VPCID:     vpcID,
		}
		ingressSecurityGroup, err = s.createSecurityGroup(ingressSGInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create security group '%s': '%#v'", ingressSGInput.GroupName, err))
		}
		ingressSecurityGroupID, err = ingressSecurityGroup.GetID()
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
	}

	var publicRouteTable *awsresources.RouteTable
	var publicSubnet *awsresources.Subnet
	var publicSubnetID string

	if !keyv2.UseCloudFormation(cluster) {
		// Create public route table.
		publicRouteTable = &awsresources.RouteTable{
			Name:   keyv2.ClusterID(cluster),
			VpcID:  vpcID,
			Client: clients.EC2,
			Logger: s.logger,
		}
		publicRouteTableCreated, err := publicRouteTable.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create route table: '%#v'", err))
		}
		if publicRouteTableCreated {
			s.logger.Log("info", "created public route table")
		} else {
			s.logger.Log("info", "route table already exists, reusing")
		}

		if err := publicRouteTable.MakePublic(); err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not make route table public: '%#v'", err))
		}

		// Create public subnet.
		subnetInput := SubnetInput{
			Name:       keyv2.SubnetName(cluster, suffixPublic),
			CidrBlock:  cluster.Spec.AWS.VPC.PublicSubnetCIDR,
			Clients:    clients,
			Cluster:    cluster,
			MakePublic: true,
			RouteTable: publicRouteTable,
			VpcID:      vpcID,
		}
		publicSubnet, err = s.createSubnet(subnetInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create public subnet: '%#v'", err))
		}

		publicSubnetID, err = publicSubnet.GetID()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get public subnet ID: '%#v'", err))
		}
	}

	var privateRouteTable *awsresources.RouteTable
	var privateSubnet *awsresources.Subnet
	var privateSubnetID string

	// For new clusters create a NAT gateway, private route table and private subnet.
	if keyv2.HasClusterVersion(cluster) && !keyv2.UseCloudFormation(cluster) {
		// Create private route table.
		privateRouteTable = &awsresources.RouteTable{
			Name:   keyv2.RouteTableName(cluster, suffixPrivate),
			VpcID:  vpcID,
			Client: clients.EC2,
			Logger: s.logger,
		}
		privateRouteTableCreated, err := privateRouteTable.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create route table: '%#v'", err))
		}
		if privateRouteTableCreated {
			s.logger.Log("info", "created private route table")
		} else {
			s.logger.Log("info", "private route table already exists, reusing")
		}

		// Create private subnet.
		subnetInput := SubnetInput{
			Name:       keyv2.SubnetName(cluster, suffixPrivate),
			CidrBlock:  cluster.Spec.AWS.VPC.PrivateSubnetCIDR,
			Clients:    clients,
			Cluster:    cluster,
			MakePublic: false,
			RouteTable: privateRouteTable,
			VpcID:      vpcID,
		}
		privateSubnet, err = s.createSubnet(subnetInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create private subnet: '%#v'", err))
		}

		privateSubnetID, err = privateSubnet.GetID()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get private subnet ID: '%#v'", err))
		}

		// Create NAT gateway.
		natGateway := &awsresources.NatGateway{
			Name:   keyv2.ClusterID(cluster),
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
			s.logger.Log("info", fmt.Sprintf("created nat gateway for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("nat gateway for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}

		natGatewayID, err := natGateway.GetID()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get nat gateway id: '%#v'", err))
		}

		// Create default route for the NAT gateway in the private route table.
		natGatewayRouteCreated, err := privateRouteTable.CreateNatGatewayRoute(natGatewayID)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create route for nat gateway: '%#v'", err))
		}
		if natGatewayRouteCreated {
			s.logger.Log("info", "created nat gateway route")
		} else {
			s.logger.Log("info", "nat gateway route already exists, reusing")
		}
	}

	if !keyv2.UseCloudFormation(cluster) {
		hostClusterRoute := &awsresources.Route{
			RouteTable:             *publicRouteTable,
			DestinationCidrBlock:   *conn.AccepterVpcInfo.CidrBlock,
			VpcPeeringConnectionID: *conn.VpcPeeringConnectionId,
			AWSEntity:              awsresources.AWSEntity{Clients: clients},
		}

		if keyv2.HasClusterVersion(cluster) {
			// New clusters have private IPs so use the private route table.
			hostClusterRoute.RouteTable = *privateRouteTable
		} else {
			// Legacy clusters have public IPs so use the public route table.
			hostClusterRoute.RouteTable = *publicRouteTable
		}

		hostRouteCreated, err := hostClusterRoute.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not add host vpc route: '%#v'", err))
		}
		if hostRouteCreated {
			s.logger.Log("info", fmt.Sprintf("created host vpc route for cluster '%s'", keyv2.ClusterID(cluster)))
		} else {
			s.logger.Log("info", fmt.Sprintf("host vpc route for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
		}

		for _, privateRouteTableName := range cluster.Spec.AWS.VPC.RouteTableNames {
			privateRouteTable := &awsresources.RouteTable{
				Name:   privateRouteTableName,
				VpcID:  cluster.Spec.AWS.VPC.PeerID,
				Client: hostClients.EC2,
				Logger: s.logger,
			}

			privateRoute := &awsresources.Route{
				RouteTable:             *privateRouteTable,
				DestinationCidrBlock:   *conn.RequesterVpcInfo.CidrBlock,
				VpcPeeringConnectionID: *conn.VpcPeeringConnectionId,
				AWSEntity:              awsresources.AWSEntity{Clients: hostClients},
			}

			privateRouteCreated, err := privateRoute.CreateIfNotExists()
			if err != nil {
				return microerror.Maskf(executionFailedError, fmt.Sprintf("could not add guest vpc route: '%#v'", err))
			}
			if privateRouteCreated {
				s.logger.Log("info", fmt.Sprintf("created guest vpc route for cluster '%s'", keyv2.ClusterID(cluster)))
			} else {
				s.logger.Log("info", fmt.Sprintf("host vpc guest for cluster '%s' already exists, reusing", keyv2.ClusterID(cluster)))
			}

		}
	}

	var apiLB *awsresources.ELB
	var etcdLB *awsresources.ELB
	var ingressLB *awsresources.ELB
	var anyMastersCreated bool
	var masterIDs []string

	if !keyv2.UseCloudFormation(cluster) {
		mastersInput := runMachinesInput{
			clients:             clients,
			cluster:             cluster,
			tlsAssets:           tlsAssets,
			clusterKeys:         clusterKeys,
			clusterName:         keyv2.ClusterID(cluster),
			bucket:              bucket,
			securityGroup:       mastersSecurityGroup,
			instanceProfileName: masterPolicy.GetName(),
			prefix:              prefixMaster,
		}

		if keyv2.HasClusterVersion(cluster) {
			// New clusters have masters in the private subnet.
			mastersInput.subnet = privateSubnet
		} else {
			// Legacy clusters have masters in the public subnet.
			mastersInput.subnet = publicSubnet

			// An EC2 Keypair is needed for legacy clusters. New clusters provide SSH keys via cloud config.
			mastersInput.keyPairName = keyv2.ClusterID(cluster)
		}

		// Run masters.
		anyMastersCreated, masterIDs, err = s.runMachines(mastersInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not start masters: '%#v'", err))
		}

		if !validateIDs(masterIDs) {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("master nodes had invalid instance IDs: %v", masterIDs))
		}

		// Create apiserver load balancer.
		lbInput := LoadBalancerInput{
			Name:               cluster.Spec.Cluster.Kubernetes.API.Domain,
			Clients:            clients,
			Cluster:            cluster,
			IdleTimeoutSeconds: cluster.Spec.AWS.API.ELB.IdleTimeoutSeconds,
			InstanceIDs:        masterIDs,
			PortsToOpen: awsresources.PortPairs{
				{
					PortELB:      cluster.Spec.Cluster.Kubernetes.API.SecurePort,
					PortInstance: cluster.Spec.Cluster.Kubernetes.API.SecurePort,
				},
			},
			SecurityGroupID: mastersSecurityGroupID,
			SubnetID:        publicSubnetID,
		}

		apiLB, err = s.createLoadBalancer(lbInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create apiserver load balancer: '%#v'", err))
		}

		// Create etcd load balancer.
		lbInput = LoadBalancerInput{
			Name:               cluster.Spec.Cluster.Etcd.Domain,
			Clients:            clients,
			Cluster:            cluster,
			IdleTimeoutSeconds: cluster.Spec.AWS.Etcd.ELB.IdleTimeoutSeconds,
			InstanceIDs:        masterIDs,
			PortsToOpen: awsresources.PortPairs{
				{
					PortELB:      httpsPort,
					PortInstance: cluster.Spec.Cluster.Etcd.Port,
				},
			},
			SecurityGroupID: mastersSecurityGroupID,
			SubnetID:        publicSubnetID,
		}

		if keyv2.HasClusterVersion(cluster) {
			lbInput.Scheme = "internal"
		}

		etcdLB, err = s.createLoadBalancer(lbInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create etcd load balancer: '%#v'", err))
		}

		// Masters were created but the master IAM policy existed from a previous
		// execution. Its likely that previous execution failed. IAM policies can't
		// be reused for EC2 instances.
		if anyMastersCreated && !masterPolicyCreated {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("cluster '%s' cannot be processed. As IAM policy for master nodes cannot be reused. Please delete this cluster.", keyv2.ClusterID(cluster)))
		}

		// Create Ingress load balancer.
		ingressLbInput := LoadBalancerInput{
			Name:               cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
			Clients:            clients,
			Cluster:            cluster,
			IdleTimeoutSeconds: cluster.Spec.AWS.Ingress.ELB.IdleTimeoutSeconds,
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

		ingressLB, err = s.createLoadBalancer(ingressLbInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create ingress load balancer: '%#v'", err))
		}

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

		// For new clusters don't assign public IPs and use the private subnet.
		if keyv2.HasClusterVersion(cluster) {
			lcInput.associatePublicIP = false
			lcInput.subnet = privateSubnet
		} else {
			lcInput.associatePublicIP = true
			lcInput.subnet = publicSubnet
		}

		// An EC2 Keypair is needed for legacy clusters. New clusters provide SSH keys via cloud config.
		if !keyv2.HasClusterVersion(cluster) {
			lcInput.keypairName = keyv2.ClusterID(cluster)
		}

		lcCreated, err := s.createLaunchConfiguration(lcInput)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create launch config: '%#v'", err))
		}

		if lcCreated {
			s.logger.Log("info", fmt.Sprintf("created worker launch config"))
		} else {
			s.logger.Log("info", fmt.Sprintf("launch config %s already exists, reusing", keyv2.ClusterID(cluster)))
		}

		workersLCName, err := launchConfigurationName(cluster, "worker", workersSecurityGroupID)
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not get launch config name: '%#v'", err))
		}

		asg := awsresources.AutoScalingGroup{
			Client:                  clients.AutoScaling,
			Name:                    keyv2.AutoScalingGroupName(cluster, prefixWorker),
			ClusterID:               keyv2.ClusterID(cluster),
			MinSize:                 keyv2.WorkerCount(cluster),
			MaxSize:                 keyv2.WorkerCount(cluster),
			AvailabilityZone:        keyv2.AvailabilityZone(cluster),
			LaunchConfigurationName: workersLCName,
			LoadBalancerName:        ingressLB.Name,
			VPCZoneIdentifier:       publicSubnetID,
			HealthCheckGracePeriod:  gracePeriodSeconds,
		}

		// For new clusters launch the workers in the private subnet.
		if keyv2.HasClusterVersion(cluster) {
			asg.VPCZoneIdentifier = privateSubnetID
		} else {
			asg.VPCZoneIdentifier = publicSubnetID
		}

		asgCreated, err := asg.CreateIfNotExists()
		if err != nil {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("could not create auto scaling group: '%#v'", err))
		}

		if asgCreated {
			s.logger.Log("info", fmt.Sprintf("created auto scaling group '%s' with size %v", asg.Name, keyv2.WorkerCount(cluster)))
		} else {
			// If the cluster exists set the worker count so the cluster can be scaled.
			scaleWorkers := awsresources.AutoScalingGroup{
				Client:  clients.AutoScaling,
				Name:    keyv2.AutoScalingGroupName(cluster, prefixWorker),
				MinSize: keyv2.WorkerCount(cluster),
				MaxSize: keyv2.WorkerCount(cluster),
			}

			if err := scaleWorkers.Update(); err != nil {
				s.logger.Log("error", fmt.Sprintf("%#v", err))
			}

			s.logger.Log("info", fmt.Sprintf("auto scaling group '%s' already exists, setting to %d workers", scaleWorkers.Name, scaleWorkers.MaxSize))
		}
	}

	// Create Record Sets for the Load Balancers.
	// During Cloud Formation migration this logic is only needed for non CF clusters.
	if !keyv2.UseCloudFormation(cluster) {
		recordSetInputs := []recordSetInput{
			{
				Cluster:      cluster,
				Client:       clients.Route53,
				Resource:     apiLB,
				Domain:       cluster.Spec.Cluster.Kubernetes.API.Domain,
				HostedZoneID: cluster.Spec.AWS.API.HostedZones,
				Type:         route53.RRTypeA,
			},
			{
				Cluster:      cluster,
				Client:       clients.Route53,
				Resource:     etcdLB,
				Domain:       cluster.Spec.Cluster.Etcd.Domain,
				HostedZoneID: cluster.Spec.AWS.Etcd.HostedZones,
				Type:         route53.RRTypeA,
			},
			{
				Cluster:      cluster,
				Client:       clients.Route53,
				Resource:     ingressLB,
				Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
				HostedZoneID: cluster.Spec.AWS.Ingress.HostedZones,
				Type:         route53.RRTypeA,
			},
			{
				Cluster:      cluster,
				Client:       clients.Route53,
				Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.WildcardDomain,
				HostedZoneID: cluster.Spec.AWS.Ingress.HostedZones,
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
	}

	if !keyv2.UseCloudFormation(cluster) {
		masterServiceInput := MasterServiceInput{
			Clients:  clients,
			Cluster:  cluster,
			MasterID: masterIDs[0],
		}
		if err := s.createMasterService(masterServiceInput); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (s *Resource) processDelete(cluster v1alpha1.AWSConfig) error {
	if err := validateCluster(cluster); err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("cluster spec is invalid: '%#v'", err))
	}

	// For new clusters using Cloud Formation there is an OperatorKit resource
	// for the k8s namespace.
	if !keyv2.UseCloudFormation(cluster) {
		if err := s.deleteClusterNamespace(cluster.Spec.Cluster); err != nil {
			s.logger.Log("error", "could not delete cluster namespace:", err)
		}
	}

	clients := awsutil.NewClients(s.awsConfig)
	err := s.awsConfig.SetAccountID(clients.IAM)
	if err != nil {
		return microerror.Maskf(executionFailedError, fmt.Sprintf("could not retrieve amazon account id: '%#v'", err))
	}

	// Retrieve AWS host cluster client.
	s.awsHostConfig.Region = cluster.Spec.AWS.Region
	hostClients := awsutil.NewClients(s.awsHostConfig)
	if err := s.awsHostConfig.SetAccountID(hostClients.IAM); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not retrieve host amazon account id: '%#v'", err))
	}

	if !keyv2.UseCloudFormation(cluster) {
		// Delete masters.
		s.logger.Log("info", "deleting masters...")
		if err := s.deleteMachines(deleteMachinesInput{
			clients:     clients,
			clusterName: keyv2.ClusterID(cluster),
			prefix:      prefixMaster,
		}); err != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", err))
		} else {
			s.logger.Log("info", "deleted masters")
		}
	}

	if keyv2.UseCloudFormation(cluster) {
		// During Cloud Formation migration we need to delete the main stack
		// so all resoures including the VPC are deleted.
		stack := awsresources.ASGStack{
			Client: clients.CloudFormation,
			Name:   keyv2.MainGuestStackName(cluster),
		}

		if err := stack.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", err))
		} else {
			s.logger.Log("info", fmt.Sprintf("deleted cloud formation stack: '%s'", stack.Name))
		}
	} else {
		// Delete workers Auto Scaling Group.
		asg := awsresources.AutoScalingGroup{
			Client: clients.AutoScaling,
			Name:   keyv2.AutoScalingGroupName(cluster, prefixWorker),
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
	}

	if !keyv2.UseCloudFormation(cluster) {
		// Delete Record Sets.

		apiLBName, err := keyv2.LoadBalancerName(cluster.Spec.Cluster.Kubernetes.API.Domain, cluster)
		etcdLBName, err := keyv2.LoadBalancerName(cluster.Spec.Cluster.Etcd.Domain, cluster)
		ingressLBName, err := keyv2.LoadBalancerName(cluster.Spec.Cluster.Kubernetes.IngressController.Domain, cluster)
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
						HostedZoneID: cluster.Spec.AWS.API.HostedZones,
						Type:         route53.RRTypeA,
					},
					{
						Cluster:      cluster,
						Client:       clients.Route53,
						Resource:     etcdLB,
						Domain:       cluster.Spec.Cluster.Etcd.Domain,
						HostedZoneID: cluster.Spec.AWS.Etcd.HostedZones,
						Type:         route53.RRTypeA,
					},
					{
						Cluster:      cluster,
						Client:       clients.Route53,
						Resource:     ingressLB,
						Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
						HostedZoneID: cluster.Spec.AWS.Ingress.HostedZones,
						Type:         route53.RRTypeA,
					},
					{
						Cluster:      cluster,
						Client:       clients.Route53,
						Value:        cluster.Spec.Cluster.Kubernetes.IngressController.Domain,
						Domain:       cluster.Spec.Cluster.Kubernetes.IngressController.WildcardDomain,
						HostedZoneID: cluster.Spec.AWS.Ingress.HostedZones,
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
	}

	if !keyv2.UseCloudFormation(cluster) {
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
		var publicRouteTable resources.ResourceWithID
		publicRouteTable = &awsresources.RouteTable{
			Name:   keyv2.ClusterID(cluster),
			Client: clients.EC2,
			Logger: s.logger,
		}
		if err := publicRouteTable.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete route table: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted route table")
		}
	}

	// Sync VPC.
	var vpc resources.ResourceWithID
	vpc = &awsresources.VPC{
		Name:      keyv2.ClusterID(cluster),
		AWSEntity: awsresources.AWSEntity{Clients: clients},
		Logger:    s.logger,
	}
	vpcID, err := vpc.GetID()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	}

	if keyv2.HasClusterVersion(cluster) && !keyv2.UseCloudFormation(cluster) {
		// Delete NAT gateway and private subnet for new clusters.
		natGateway := &awsresources.NatGateway{
			Name: keyv2.ClusterID(cluster),
			// Dependencies.
			Logger:    s.logger,
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}
		if err := natGateway.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete nat gateway: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted nat gateway")
		}

		// Delete private route table.
		privateRouteTable := &awsresources.RouteTable{
			Name:   keyv2.RouteTableName(cluster, suffixPrivate),
			Client: clients.EC2,
			Logger: s.logger,
		}
		if err := privateRouteTable.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete private route table: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted private route table")
		}

		// Delete private subnet.
		subnetInput := SubnetInput{
			Name:    keyv2.SubnetName(cluster, suffixPrivate),
			Clients: clients,
		}
		if err := s.deleteSubnet(subnetInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete private subnet: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted private subnet")
		}
	}

	if !keyv2.UseCloudFormation(cluster) {
		// Delete internet gateway.
		internetGateway := &awsresources.InternetGateway{
			Name:  keyv2.ClusterID(cluster),
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
			Name:    keyv2.SubnetName(cluster, suffixPublic),
			Clients: clients,
		}
		if err := s.deleteSubnet(subnetInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete public subnet: '%#v'", err))
		} else {
			s.logger.Log("info", "deleted public subnet")
		}
	}

	if !keyv2.UseCloudFormation(cluster) {
		// Before the security groups can be deleted any rules referencing other
		// groups must first be deleted.
		mastersSGRulesInput := securityGroupRulesInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixMaster),
		}
		if err := s.deleteSecurityGroupRules(mastersSGRulesInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete rules for security group '%s': '%#v'", mastersSGRulesInput.GroupName, err))
		}

		workersSGRulesInput := securityGroupRulesInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixWorker),
		}
		if err := s.deleteSecurityGroupRules(workersSGRulesInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete rules for security group '%s': '%#v'", mastersSGRulesInput.GroupName, err))
		}

		ingressSGRulesInput := securityGroupRulesInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixIngress),
		}
		if err := s.deleteSecurityGroupRules(ingressSGRulesInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete rules for security group '%s': '%#v'", mastersSGRulesInput.GroupName, err))
		}

		// Delete masters security group.
		mastersSGInput := securityGroupInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixMaster),
		}
		if err := s.deleteSecurityGroup(mastersSGInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': '%#v'", mastersSGInput.GroupName, err))
		}

		// Delete workers security group.
		workersSGInput := securityGroupInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixWorker),
		}
		if err := s.deleteSecurityGroup(workersSGInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': '%#v'", workersSGInput.GroupName, err))
		}

		// Delete ingress security group.
		ingressSGInput := securityGroupInput{
			Clients:   clients,
			GroupName: keyv2.SecurityGroupName(cluster, prefixIngress),
		}
		if err := s.deleteSecurityGroup(ingressSGInput); err != nil {
			s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': '%#v'", ingressSGInput.GroupName, err))
		}
	}

	if !keyv2.UseCloudFormation(cluster) {
		vpcPeeringConection := &awsresources.VPCPeeringConnection{
			VPCId:     vpcID,
			PeerVPCId: cluster.Spec.AWS.VPC.PeerID,
			AWSEntity: awsresources.AWSEntity{
				Clients:     clients,
				HostClients: hostClients,
			},
			Logger: s.logger,
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
				Logger: s.logger,
			}

			privateRoute := &awsresources.Route{
				RouteTable:             *privateRouteTable,
				DestinationCidrBlock:   *conn.RequesterVpcInfo.CidrBlock,
				VpcPeeringConnectionID: *conn.VpcPeeringConnectionId,
				AWSEntity:              awsresources.AWSEntity{Clients: hostClients},
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
			ClusterID:  keyv2.ClusterID(cluster),
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
			ClusterID:  keyv2.ClusterID(cluster),
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
			Name:      keyv2.ClusterID(cluster),
			AWSEntity: awsresources.AWSEntity{Clients: clients},
		}
		if err := kmsKey.Delete(); err != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", err))
		} else {
			s.logger.Log("info", "deleted KMS key")
		}
	}

	// Delete keypair.
	var keyPair resources.Resource
	keyPair = &awsresources.KeyPair{
		ClusterName: keyv2.ClusterID(cluster),
		AWSEntity:   awsresources.AWSEntity{Clients: clients},
	}
	if err := keyPair.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err))
	} else {
		s.logger.Log("info", "deleted keypair")
	}

	return nil
}

func (s *Resource) uploadCloudconfigToS3(svc *s3.S3, s3Bucket, path, data string) error {
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

func (s *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	// currently only used for cloudformation
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	if keyv2.UseCloudFormation(customObject) {
		s.logger.LogCtx(ctx, "debug", "looking for AWS stack")

		stackName := keyv2.MainGuestStackName(customObject)

		describeInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}
		describeOutput, err := s.awsClients.CloudFormation.DescribeStacks(describeInput)

		if IsStackNotFound(err) {
			s.logger.LogCtx(ctx, "debug", "did not find a stack in AWS API")
			return StackState{}, nil
		}
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		if len(describeOutput.Stacks) > 1 {
			return StackState{}, microerror.Mask(notFoundError)
		}

		// current is called on cluster deletion, if the stack creation failed the
		// outputs can be unaccessible, this can lead to a stack that cannot be deleted.
		// it can also be called during creation, while the outputs are still not
		// accessible.
		status := describeOutput.Stacks[0].StackStatus
		errorStatuses := []string{
			"ROLLBACK_IN_PROGRESS",
			"ROLLBACK_COMPLETE",
			"CREATE_IN_PROGRESS",
		}
		for _, errorStatus := range errorStatuses {
			if *status == errorStatus {
				outputStackState := StackState{
					Name:           stackName,
					Workers:        "",
					ImageID:        "",
					ClusterVersion: "",
				}
				return outputStackState, nil
			}
		}

		outputs := describeOutput.Stacks[0].Outputs

		workers, err := getStackOutputValue(outputs, workersOutputKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		imageID, err := getStackOutputValue(outputs, imageIDOutputKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		clusterVersion, err := getStackOutputValue(outputs, clusterVersionOutputKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		outputStackState := StackState{
			Name:           stackName,
			Workers:        workers,
			ImageID:        imageID,
			ClusterVersion: clusterVersion,
		}

		return outputStackState, nil
	}
	return nil, nil
}

func (s *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	// currently only used for cloudformation
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if keyv2.UseCloudFormation(customObject) {
		mainStack, err := newMainStack(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return mainStack, nil
	}
	return nil, nil
}

func (s *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cluster, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	s.logger.Log("info", fmt.Sprintf("creating cluster '%s'", keyv2.ClusterID(cluster)))

	// cloudformation logic: on creation we only need to handle the cloudformation resource, legacy is handled by the update methods
	// Once the transition to cloudformation is done we will separate the cloudformation and legacy resources.
	if keyv2.UseCloudFormation(cluster) {
		stackInput, err := toCreateStackInput(createChange)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = s.awsClients.CloudFormation.CreateStack(&stackInput)
		if err != nil {
			return microerror.Mask(err)
		}
		err = s.awsHostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: stackInput.StackName,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "debug", "creating AWS cloudformation stack: created")

		// Create host post-main stack. It includes the peering routes, which need resources from the
		// guest stack to be in place before it can be created.
		err = s.createHostPostStack(cluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}

func (s *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	s.logger.Log("info", "in ApplyDeleteChange ")
	return nil
}

func (s *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func toStackState(v interface{}) (StackState, error) {
	if v == nil {
		return StackState{}, nil
	}

	stackState, ok := v.(StackState)
	if !ok {
		return StackState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", stackState, v)
	}

	return stackState, nil
}

func toUpdateStackInput(v interface{}) (cloudformation.UpdateStackInput, error) {
	if v == nil {
		return cloudformation.UpdateStackInput{}, nil
	}

	updateStackInput, ok := v.(cloudformation.UpdateStackInput)
	if !ok {
		return cloudformation.UpdateStackInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", updateStackInput, v)
	}

	return updateStackInput, nil
}

func (s *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	currentStackState, err := toStackState(currentState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "debug", "finding out if the main stack should be updated")

	updateState := cloudformation.UpdateStackInput{
		StackName: aws.String(""),
	}

	if currentStackState.Name != "" && !reflect.DeepEqual(desiredStackState, currentStackState) {
		s.logger.LogCtx(ctx, "debug", "main stack should be updated")
		var mainTemplate string
		mainTemplate, err := s.getMainGuestTemplateBody(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		updateState.StackName = aws.String(desiredStackState.Name)
		updateState.TemplateBody = aws.String(mainTemplate)
	}

	return updateState, nil
}

func (s *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "debug", "finding out if the main stack should be created")

	createState := cloudformation.CreateStackInput{
		StackName: aws.String(""),
	}

	if desiredStackState.Name != "" {
		s.logger.LogCtx(ctx, "debug", "main stack should be created")

		// We need to create the required peering resources in the host account before
		// getting the guest main stack template body, it requires id values from host
		// resources.
		err = s.createHostPreStack(customObject)
		if err != nil {
			return cloudformation.CreateStackInput{}, microerror.Mask(err)
		}

		var mainTemplate string
		mainTemplate, err := s.getMainGuestTemplateBody(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		createState.StackName = aws.String(desiredStackState.Name)
		createState.TemplateBody = aws.String(mainTemplate)
		createState.TimeoutInMinutes = aws.Int64(defaultCreationTimeout)
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		createState.Capabilities = []*string{
			aws.String("CAPABILITY_NAMED_IAM"),
		}
	}

	return createState, nil
}

func getStackOutputValue(outputs []*cloudformation.Output, key string) (string, error) {
	for _, o := range outputs {
		if *o.OutputKey == key {
			return *o.OutputValue, nil
		}
	}

	return "", microerror.Mask(notFoundError)
}

func toCreateStackInput(v interface{}) (cloudformation.CreateStackInput, error) {
	if v == nil {
		return cloudformation.CreateStackInput{}, nil
	}

	createStackInput, ok := v.(cloudformation.CreateStackInput)
	if !ok {
		return cloudformation.CreateStackInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", createStackInput, v)
	}

	return createStackInput, nil
}

func (s *Resource) createHostPreStack(customObject v1alpha1.AWSConfig) error {
	stackName := keyv2.MainHostPreStackName(customObject)
	mainTemplate, err := s.getMainHostPreTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		Capabilities: []*string{
			aws.String("CAPABILITY_NAMED_IAM"),
		},
	}

	s.logger.Log("debug", "creating AWS Host Pre-Guest cloudformation stack")
	_, err = s.awsHostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.awsHostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	s.logger.Log("debug", "creating AWS Host Pre-Guest cloudformation stack: created")
	return nil
}

func (s *Resource) createHostPostStack(customObject v1alpha1.AWSConfig) error {
	stackName := keyv2.MainHostPostStackName(customObject)
	mainTemplate, err := s.getMainHostPostTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
	}

	s.logger.Log("debug", "creating AWS Host Post-Guest cloudformation stack")
	_, err = s.awsHostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.awsHostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	s.logger.Log("debug", "creating AWS Host Post-Guest cloudformation stack: created")
	return nil
}
