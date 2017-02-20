package create

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/spec/node"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
)

const (
	ClusterListAPIEndpoint  string = "/apis/cluster.giantswarm.io/v1/awses"
	ClusterWatchAPIEndpoint string = "/apis/cluster.giantswarm.io/v1/watch/awses"
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

const (
	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#InstanceState
	EC2TerminatedState = 48
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Dependencies.
	AwsSession *awssession.Session
	EC2Client  *ec2.EC2
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		AwsSession: nil,
		EC2Client:  nil,
		K8sClient:  nil,
		Logger:     nil,
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
		awsSession: config.AwsSession,
		ec2Client:  config.EC2Client,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	awsSession *awssession.Session
	ec2Client  *ec2.EC2
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

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

					if err := s.createClusterNamespace(*cluster); err != nil {
						s.logger.Log("error", fmt.Sprintf("could not create cluster namespace: %s", err))
						return
					}

					// Run masters
					if err := s.runMachines(cluster.Spec.Masters, cluster.Name, "master"); err != nil {
						s.logger.Log("error", microerror.MaskAny(err))
						return
					}

					// Run workers
					if err := s.runMachines(cluster.Spec.Workers, cluster.Name, "worker"); err != nil {
						s.logger.Log("error", microerror.MaskAny(err))
						return
					}

					s.logger.Log("info", fmt.Sprintf("cluster '%s' processed", cluster.Name))
				},
				DeleteFunc: func(obj interface{}) {
					cluster := obj.(*awstpr.CustomObject)
					s.logger.Log("info", fmt.Sprintf("cluster '%s' deleted", cluster.Name))

					if err := s.deleteClusterNamespace(*cluster); err != nil {
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

func (s *Service) runMachines(machines []node.Node, clusterName, prefix string) error {
	for no, machine := range machines {
		name := fmt.Sprintf("%s-%d", prefix, no)
		if err := s.runMachine(machine, clusterName, name); err != nil {
			return microerror.MaskAny(err)
		}
	}
	return nil
}

func (s *Service) runMachine(machine node.Node, clusterName, name string) error {
	instances, err := s.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(name),
				},
			},
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyCluster)),
				Values: []*string{
					aws.String(clusterName),
				},
			},
		},
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	// If the instance doesn't exist, then the Reservation field should be nil.
	// Otherwise, it will contain a slice of instances (which is going to contain our one instance we queried for).
	// TODO(nhlfr): Check whether the instance has correct parameters. That will be most probably done when we
	// will introduce the interface for creating, deleting and updating resources.
	if instances.Reservations != nil {
		for _, r := range instances.Reservations {
			for _, i := range r.Instances {
				if *i.State.Code != EC2TerminatedState {
					s.logger.Log("info", fmt.Sprintf("instance '%s' already exists", name))
					return nil
				}
			}
		}
	}

	reservation, err := s.ec2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(machine.ImageID),
		InstanceType: aws.String(machine.InstanceType),
		MinCount:     aws.Int64(int64(1)),
		MaxCount:     aws.Int64(int64(1)),
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("instance '%s' reserved", name))

	if _, err := s.ec2Client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{reservation.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(name),
			},
			{
				Key:   aws.String(tagKeyCluster),
				Value: aws.String(clusterName),
			},
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("instance '%s' tagged", name))

	return nil
}
