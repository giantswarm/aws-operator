package create

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/awstpr"
	awsinfo "github.com/giantswarm/awstpr/aws"
	"github.com/giantswarm/clustertpr/node"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
)

const (
	ClusterListAPIEndpoint  string = "/apis/cluster.giantswarm.io/v1/awses"
	ClusterWatchAPIEndpoint string = "/apis/cluster.giantswarm.io/v1/watch/awses"
	// The format of instance's name is "[name of cluster]-[prefix ('master' or 'worker')]-[number]".
	instanceNameFormat string = "%s-%s-%d"
	// Period or re-synchronizing the list of objects in k8s watcher. 0 means that re-sync will be
	// delayed as long as possible, until the watch will be closed or timed out.
	resyncPeriod time.Duration = 0
	// Prefixes used for machine names.
	prefixMaster string = "master"
	prefixWorker string = "worker"
	// EC2 instance tag keys.
	tagKeyName    string = "Name"
	tagKeyCluster string = "Cluster"
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
	// Dependencies.
	AwsConfig awsutil.Config
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
	CertsDir  string
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
		CertsDir:  "",
	}
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		awsConfig: config.AwsConfig,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		// AWS certificates options.
		certsDir: config.CertsDir,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	awsConfig awsutil.Config
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// AWS certificates options.
	certsDir string

	// Internals.
	bootOnce sync.Once
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
			stream, err := req.Stream()
			if err != nil {
				return nil, err
			}

			watcher := watch.NewStreamWatcher(&k8sutil.ClusterDecoder{
				Stream: stream,
			})

			return watcher, nil
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
			resyncPeriod,
			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					cluster := obj.(*awstpr.CustomObject)
					s.logger.Log("info", fmt.Sprintf("creating cluster '%s'", cluster.Name))

					if err := s.createClusterNamespace(cluster.Spec.Cluster); err != nil {
						s.logger.Log("error", fmt.Sprintf("could not create cluster namespace: %s", err))
						return
					}

					// Create AWS client
					s.awsConfig.Region = cluster.Spec.AWS.Region
					awsSession, ec2Client := awsutil.NewClient(s.awsConfig)

					// Create KMS key
					kmsSvc := kms.New(awsSession)
					key, err := kmsSvc.CreateKey(&kms.CreateKeyInput{})
					if err != nil {
						s.logger.Log("error", fmt.Sprintf("could not create KMS service client: %s", err))
						return
					}

					// Encode TLS assets
					tlsAssets, err := s.encodeTLSAssets(awsSession, *key.KeyMetadata.Arn)
					if err != nil {
						s.logger.Log("error", fmt.Sprintf("could not encode TLS assets: %s", err))
						return
					}

					// Run masters
					if err := s.runMachines(runMachinesInput{
						awsSession:  awsSession,
						ec2Client:   ec2Client,
						spec:        cluster.Spec,
						tlsAssets:   tlsAssets,
						clusterName: cluster.Name,
						prefix:      prefixMaster,
					}); err != nil {
						s.logger.Log("error", microerror.MaskAny(err))
						return
					}

					// Run workers
					if err := s.runMachines(runMachinesInput{
						awsSession:  awsSession,
						ec2Client:   ec2Client,
						spec:        cluster.Spec,
						tlsAssets:   tlsAssets,
						clusterName: cluster.Name,
						prefix:      prefixWorker,
					}); err != nil {
						s.logger.Log("error", microerror.MaskAny(err))
						return
					}

					s.logger.Log("info", fmt.Sprintf("cluster '%s' processed", cluster.Name))
				},
				DeleteFunc: func(obj interface{}) {
					cluster := obj.(*awstpr.CustomObject)
					s.logger.Log("info", fmt.Sprintf("cluster '%s' deleted", cluster.Name))

					if err := s.deleteClusterNamespace(cluster.Spec.Cluster); err != nil {
						s.logger.Log("error", "could not delete cluster namespace:", err)
					}
				},
			},
		)

		s.logger.Log("info", "starting watch")

		// Cluster informer lifecycle can be interrupted by putting a value into a "stop channel".
		// We aren't currently using that functionality, so we are passing a nil here.
		clusterInformer.Run(nil)
	})
}

type runMachinesInput struct {
	awsSession  *awssession.Session
	ec2Client   *ec2.EC2
	spec        awstpr.Spec
	tlsAssets   *cloudconfig.CompactTLSAssets
	clusterName string
	prefix      string
}

func (s *Service) runMachines(input runMachinesInput) error {
	var (
		machines    []node.Node
		awsMachines []awsinfo.Node
	)

	switch input.prefix {
	case prefixMaster:
		machines = input.spec.Cluster.Masters
		awsMachines = input.spec.AWS.Masters
	case prefixWorker:
		machines = input.spec.Cluster.Workers
		awsMachines = input.spec.AWS.Workers
	}

	// TODO(nhlfr): Create a separate module for validating specs and execute on the earlier stages.
	if len(machines) != len(awsMachines) {
		return microerror.MaskAny(fmt.Errorf("mismatched number of %s machines in the 'spec' and 'aws' sections: %d != %d",
			input.prefix,
			len(machines),
			len(awsMachines)))
	}

	for i := 0; i < len(machines); i++ {
		name := fmt.Sprintf(instanceNameFormat, input.clusterName, input.prefix, i)
		if err := s.runMachine(runMachineInput{
			awsSession:  input.awsSession,
			ec2Client:   input.ec2Client,
			spec:        input.spec,
			machine:     machines[i],
			awsNode:     awsMachines[i],
			tlsAssets:   input.tlsAssets,
			clusterName: input.clusterName,
			name:        name,
			prefix:      input.prefix,
		}); err != nil {
			return microerror.MaskAny(err)
		}
	}
	return nil
}

func allExistingInstancesMatch(instances *ec2.DescribeInstancesOutput, state EC2StateCode) bool {
	// If the instance doesn't exist, then the Reservation field should be nil.
	// Otherwise, it will contain a slice of instances (which is going to contain our one instance we queried for).
	// TODO(nhlfr): Check whether the instance has correct parameters. That will be most probably done when we
	// will introduce the interface for creating, deleting and updating resources.
	if instances.Reservations != nil {
		for _, r := range instances.Reservations {
			for _, i := range r.Instances {
				if *i.State.Code != int64(state) {
					return false
				}
			}
		}
	}
	return true
}

type runMachineInput struct {
	awsSession  *awssession.Session
	ec2Client   *ec2.EC2
	spec        awstpr.Spec
	machine     node.Node
	awsNode     awsinfo.Node
	tlsAssets   *cloudconfig.CompactTLSAssets
	clusterName string
	name        string
	prefix      string
}

func (s *Service) runMachine(input runMachineInput) error {
	instances, err := input.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(input.name),
				},
			},
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyCluster)),
				Values: []*string{
					aws.String(input.clusterName),
				},
			},
		},
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	cloudConfigParams := cloudconfig.CloudConfigTemplateParams{
		Cluster:   input.spec.Cluster,
		Node:      input.machine,
		TLSAssets: *input.tlsAssets,
	}

	cloudConfig, err := s.cloudConfig(input.prefix, cloudConfigParams)
	if err != nil {
		return err
	}

	if !allExistingInstancesMatch(instances, EC2TerminatedState) {
		s.logger.Log("info", fmt.Sprintf("instance '%s' already exists", input.name))
		return nil
	}

	reservation, err := input.ec2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(input.awsNode.ImageID),
		InstanceType: aws.String(input.awsNode.InstanceType),
		MinCount:     aws.Int64(int64(1)),
		MaxCount:     aws.Int64(int64(1)),
		UserData:     aws.String(cloudConfig),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(ProfileName),
		},
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("instance '%s' reserved", input.name))

	if _, err := input.ec2Client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{reservation.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(input.name),
			},
			{
				Key:   aws.String(tagKeyCluster),
				Value: aws.String(input.clusterName),
			},
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("instance '%s' tagged", input.name))

	return nil
}
