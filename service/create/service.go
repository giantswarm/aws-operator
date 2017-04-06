package create

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/awstpr"
	awsinfo "github.com/giantswarm/awstpr/aws"
	"github.com/giantswarm/clustertpr/node"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
	// micrologger "github.com/giantswarm/microkit/logger"
	"github.com/juju/errgo"
	"k8s.io/client-go/tools/cache"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/common"
)

const (
	// The format of instance's name is "[name of cluster]-[prefix ('master' or 'worker')]-[number]".
	instanceNameFormat string = "%s-%s-%d"
	// The format of prefix inside a cluster "[name of cluster]-[prefix ('master' or 'worker')]".
	instanceClusterPrefixFormat string = "%s-%s"
	// Period or re-synchronizing the list of objects in k8s watcher. 0 means that re-sync will be
	// delayed as long as possible, until the watch will be closed or timed out.
	resyncPeriod time.Duration = 0
	// Prefixes used for machine names.
	prefixMaster string = "master"
	prefixWorker string = "worker"
	// EC2 instance tag keys.
	tagKeyName    string = "Name"
	tagKeyCluster string = "Cluster"
	// Number of retries of RunInstances to wait for Roles to propagate to
	// Instance Profiles
	runInstancesRetries = 10
)

type EC2StateCode int

const (
	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#InstanceState
	EC2PendingState      EC2StateCode = 0
	EC2RunningState      EC2StateCode = 16
	EC2ShuttingDownState EC2StateCode = 32
	EC2TerminatedState   EC2StateCode = 48
	EC2StoppingState     EC2StateCode = 64
	EC2StoppedState      EC2StateCode = 80
)

// Config represents the configuration used to create a version service.
type Config struct {
	PubKeyFile string
	common.Config
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "Logger must not be empty")
	}

	newService := &Service{
		// AWS certificates options.
		pubKeyFile: config.PubKeyFile,

		Service: common.Service{
			// Dependencies.
			AwsConfig: config.AwsConfig,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			// Internals
			BootOnce: sync.Once{},
		},
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// AWS certificates options.
	pubKeyFile string

	common.Service
}

func (s *Service) Boot() {
	s.BootOnce.Do(func() {
		if err := s.createTPR(); err != nil {
			panic(err)
		}
		s.Logger.Log("info", "successfully created third-party resource")

		_, clusterInformer := cache.NewInformer(
			s.NewClusterListWatch(),
			&awstpr.CustomObject{},
			resyncPeriod,
			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					cluster := *obj.(*awstpr.CustomObject)
					s.Logger.Log("info", fmt.Sprintf("creating cluster '%s'", cluster.Name))

					if err := s.createClusterNamespace(cluster.Spec.Cluster); err != nil {
						s.Logger.Log("error", fmt.Sprintf("could not create cluster namespace: %s", errgo.Details(err)))
						return
					}

					// Create AWS client
					s.AwsConfig.Region = cluster.Spec.AWS.Region
					clients := awsutil.NewClients(s.AwsConfig)

					// Create keypair
					var keyPair resources.Resource
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
							s.Logger.Log("error", fmt.Sprintf("could not create keypair: %s", errgo.Details(err)))
							return
						}
					}

					if keyPairCreated {
						s.Logger.Log("info", fmt.Sprintf("created keypair '%s'", cluster.Name))
					} else {
						s.Logger.Log("info", fmt.Sprintf("keypair '%s' already exists, reusing", cluster.Name))
					}

					clusterID := cluster.Spec.Cluster.Cluster.ID
					certs, err := s.getCertsFromSecrets(clusterID)
					if err != nil {
						s.Logger.Log("error", fmt.Sprintf("could not get certificates from secrets: %v", errgo.Details(err)))
						return
					}

					// Create KMS key
					var kmsKey resources.ArnResource
					var kmsKeyErr error
					{
						kmsKey = &awsresources.KMSKey{
							AWSEntity: awsresources.AWSEntity{Clients: clients},
						}
						kmsKeyErr = kmsKey.CreateOrFail()
					}

					// Encode TLS assets
					tlsAssets, err := s.encodeTLSAssets(certs, clients.KMS, kmsKey.Arn())
					if err != nil {
						s.Logger.Log("error", fmt.Sprintf("could not encode TLS assets: %s", errgo.Details(err)))
						return
					}

					// Create policy
					var policy resources.NamedResource
					var policyErr error
					{
						policy = &awsresources.Policy{
							ClusterID: cluster.Spec.Cluster.Cluster.ID,
							KMSKeyArn: kmsKey.Arn(),
							S3Bucket:  awsresources.BucketName(cluster),
							AWSEntity: awsresources.AWSEntity{Clients: clients},
						}
						policyErr = policy.CreateOrFail()
					}

					// Create S3 bucket
					var bucket resources.Resource
					var bucketCreated bool
					{
						var err error
						bucket = &awsresources.Bucket{
							Name:      awsresources.BucketName(cluster),
							AWSEntity: awsresources.AWSEntity{Clients: clients},
						}
						bucketCreated, err = bucket.CreateIfNotExists()
						if err != nil {
							s.Logger.Log("error", fmt.Sprintf("could not create S3 bucket: %s", errgo.Details(err)))
							return
						}
					}

					if bucketCreated {
						s.Logger.Log("info", fmt.Sprintf("created bucket '%s'", awsresources.BucketName(cluster)))
					} else {
						s.Logger.Log("info", fmt.Sprintf("bucket '%s' already exists, reusing", awsresources.BucketName(cluster)))
					}

					// Run masters
					anyMastersCreated, masterIDs, err := s.runMachines(runMachinesInput{
						clients:             clients,
						cluster:             cluster,
						tlsAssets:           tlsAssets,
						clusterName:         cluster.Name,
						bucket:              bucket,
						keyPairName:         cluster.Name,
						instanceProfileName: policy.Name(),
						prefix:              prefixMaster,
					})
					if err != nil {
						s.Logger.Log("error", errgo.Details(err))
					}

					if !validateIDs(masterIDs) {
						s.Logger.Log("error", fmt.Sprintf("master nodes had invalid instance IDs: %v", masterIDs))
						return
					}

					// Add an elastic IP to the master
					masterID := masterIDs[0]
					s.Logger.Log("debug", fmt.Sprintf("waiting for %s to be ready", masterID))
					if err := clients.EC2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
						InstanceIds: []*string{
							aws.String(masterID),
						},
					}); err != nil {
						s.Logger.Log("error", fmt.Sprintf("master took too long to get running, aborting: %v", err))
						return
					}

					var elasticIP resources.NamedResource
					{
						elasticIP = &awsresources.ElasticIP{
							InstanceID: masterID,
							AWSEntity:  awsresources.AWSEntity{Clients: clients},
						}
						if err := elasticIP.CreateOrFail(); err != nil {
							s.Logger.Log("error", errgo.Details(err))
						}
					}
					s.Logger.Log("info", fmt.Sprintf("attached ip %v to instance %v", elasticIP.Name(), masterID))

					// Run workers
					anyWorkersCreated, _, err := s.runMachines(runMachinesInput{
						clients:             clients,
						cluster:             cluster,
						tlsAssets:           tlsAssets,
						bucket:              bucket,
						clusterName:         cluster.Name,
						keyPairName:         cluster.Name,
						instanceProfileName: policy.Name(),
						prefix:              prefixWorker,
					})
					if err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}

					// If the policy couldn't be created and some instances didn't exist before, that means that the cluster
					// is inconsistent and most problably its deployment broke in the middle during the previous run of
					// aws-operator.
					if (anyMastersCreated || anyWorkersCreated) && (kmsKeyErr != nil || policyErr != nil) {
						s.Logger.Log("error", fmt.Sprintf("cluster '%s' is inconsistent, KMS keys and policies were not created, but EC2 instances were missing, please consider deleting this cluster", cluster.Name))
						return
					}

					s.Logger.Log("info", fmt.Sprintf("cluster '%s' processed", cluster.Name))
				},
				DeleteFunc: func(obj interface{}) {
					// TODO(nhlfr): Move this to a separate operator.
					cluster := *obj.(*awstpr.CustomObject)

					if err := s.deleteClusterNamespace(cluster.Spec.Cluster); err != nil {
						s.Logger.Log("error", "could not delete cluster namespace:", err)
					}

					clients := awsutil.NewClients(s.AwsConfig)

					// Delete masters
					if err := s.deleteMachines(deleteMachinesInput{
						clients:     clients,
						clusterName: cluster.Name,
						prefix:      prefixMaster,
					}); err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}
					s.Logger.Log("info", "deleted masters")

					// Delete workers
					if err := s.deleteMachines(deleteMachinesInput{
						clients:     clients,
						clusterName: cluster.Name,
						prefix:      prefixWorker,
					}); err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}
					s.Logger.Log("info", "deleted workers")

					// Delete S3 bucket objects
					var bucket resources.Resource
					bucket = &awsresources.Bucket{
						Name:      awsresources.BucketName(cluster),
						AWSEntity: awsresources.AWSEntity{Clients: clients},
					}

					var masterBucketObject resources.Resource
					masterBucketObject = &awsresources.BucketObject{
						Name:      awsresources.BucketObjectName(cluster, prefixMaster),
						Bucket:    bucket.(*awsresources.Bucket),
						AWSEntity: awsresources.AWSEntity{Clients: clients},
					}
					if err := masterBucketObject.Delete(); err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}

					var workerBucketObject resources.Resource
					workerBucketObject = &awsresources.BucketObject{
						Name:      awsresources.BucketObjectName(cluster, prefixWorker),
						Bucket:    bucket.(*awsresources.Bucket),
						AWSEntity: awsresources.AWSEntity{Clients: clients},
					}
					if err := workerBucketObject.Delete(); err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}

					s.Logger.Log("info", "deleted bucket objects")

					// Delete policy
					var policy resources.NamedResource
					policy = &awsresources.Policy{
						ClusterID: cluster.Spec.Cluster.Cluster.ID,
						S3Bucket:  awsresources.BucketName(cluster),
						AWSEntity: awsresources.AWSEntity{Clients: clients},
					}
					if err := policy.Delete(); err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}
					s.Logger.Log("info", "deleted roles, policies, instance profiles")

					// Delete keypair
					var keyPair resources.Resource
					keyPair = &awsresources.KeyPair{
						ClusterName: cluster.Name,
						AWSEntity:   awsresources.AWSEntity{Clients: clients},
					}
					if err := keyPair.Delete(); err != nil {
						s.Logger.Log("error", errgo.Details(err))
						return
					}
					s.Logger.Log("info", "deleted keypair")

					s.Logger.Log("info", fmt.Sprintf("cluster '%s' deleted", cluster.Name))
				},
			},
		)

		s.Logger.Log("info", "starting watch")

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
	tlsAssets           *cloudconfig.CompactTLSAssets
	bucket              resources.Resource
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
			tlsAssets:           input.tlsAssets,
			bucket:              input.bucket,
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
func allExistingInstancesMatch(instances *ec2.DescribeInstancesOutput, state EC2StateCode) (*string, bool) {
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
		return microerror.MaskAny(err)
	}

	return nil
}

type runMachineInput struct {
	clients             awsutil.Clients
	cluster             awstpr.CustomObject
	machine             node.Node
	awsNode             awsinfo.Node
	tlsAssets           *cloudconfig.CompactTLSAssets
	bucket              resources.Resource
	clusterName         string
	keyPairName         string
	instanceProfileName string
	name                string
	prefix              string
}

func (s *Service) runMachine(input runMachineInput) (bool, string, error) {
	cloudConfigParams := cloudconfig.CloudConfigTemplateParams{
		Cluster:   input.cluster.Spec.Cluster,
		Node:      input.machine,
		TLSAssets: *input.tlsAssets,
	}

	cloudConfig, err := s.cloudConfig(input.prefix, cloudConfigParams, input.cluster.Spec)
	if err != nil {
		return false, "", microerror.MaskAny(err)
	}

	// We now upload the instance cloudconfig to S3 and create a "small
	// cloudconfig" that just fetches the previously uploaded "final
	// cloudconfig" and executes coreos-cloudinit with it as argument.
	// We do this to circumvent the 16KB limit on user-data for EC2 instances.
	cloudconfigConfig := SmallCloudconfigConfig{
		MachineType: input.prefix,
		Region:      input.cluster.Spec.AWS.Region,
		S3DirURI:    awsresources.BucketObjectFullDirPath(input.cluster),
	}

	var cloudconfigS3 resources.Resource
	cloudconfigS3 = &awsresources.BucketObject{
		Name:      awsresources.BucketObjectName(input.cluster, input.prefix),
		Data:      cloudConfig,
		Bucket:    input.bucket.(*awsresources.Bucket),
		AWSEntity: awsresources.AWSEntity{Clients: input.clients},
	}
	if err := cloudconfigS3.CreateOrFail(); err != nil {
		return false, "", microerror.MaskAny(err)
	}

	smallCloudconfig, err := s.SmallCloudconfig(cloudconfigConfig)
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
			AWSEntity:              awsresources.AWSEntity{Clients: input.clients},
		}
		instanceCreated, err = instance.CreateIfNotExists()
		if err != nil {
			return false, "", microerror.MaskAny(err)
		}
	}

	if instanceCreated {
		s.Logger.Log("info", fmt.Sprintf("instance '%s' reserved", input.name))
	} else {
		s.Logger.Log("info", fmt.Sprintf("instance '%s' already exists, reusing", input.name))
	}

	s.Logger.Log("info", fmt.Sprintf("instance '%s' tagged", input.name))

	return instanceCreated, instance.InstanceID, nil
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

func (s *Service) deleteMachine(input deleteMachineInput) error {
	var instance resources.Resource
	instance = &awsresources.Instance{
		Name: input.name,
	}
	if err := instance.Delete(); err != nil {
		return microerror.MaskAny(err)
	}

	s.Logger.Log("info", fmt.Sprintf("instance '%s' removed", input.name))

	return nil
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
