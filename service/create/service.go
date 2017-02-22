package create

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/awstpr"
	tpraws "github.com/giantswarm/awstpr/spec/aws"
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
	prefixMaster               = "master"
	prefixWorker               = "worker"
)

const (
	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#InstanceState
	EC2TerminatedState = 48
)

const (
	// EC2 instance tag keys.
	tagKeyName    = "Name"
	tagKeyCluster = "Cluster"
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Dependencies.
	AwsConfig             awsutil.Config
	K8sClient             kubernetes.Interface
	Logger                micrologger.Logger
	CertsDir              string
	CloudconfigMasterPath string
	CloudconfigWorkerPath string
}

// awsNode combines the generic node information of the TPR with the aws
// specific one
type awsNode struct {
	Node    node.Node
	AwsInfo tpraws.Node
}

// cloudconfigTemplateParams represents the parameters for a cloudconfig
// template for a particular node
type cloudconfigTemplateParams struct {
	Spec      awstpr.Spec
	Node      awsNode
	TLSAssets CompactTLSAssets
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		AwsConfig:             awsutil.Config{},
		K8sClient:             nil,
		Logger:                nil,
		CertsDir:              "",
		CloudconfigMasterPath: "",
		CloudconfigWorkerPath: "",
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
		awsConfig:             config.AwsConfig,
		k8sClient:             config.K8sClient,
		logger:                config.Logger,
		certsDir:              config.CertsDir,
		cloudconfigMasterPath: config.CloudconfigMasterPath,
		cloudconfigWorkerPath: config.CloudconfigWorkerPath,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	awsConfig             awsutil.Config
	k8sClient             kubernetes.Interface
	logger                micrologger.Logger
	certsDir              string
	cloudconfigMasterPath string
	cloudconfigWorkerPath string

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
					if err := s.runMachines(cluster.Spec, cluster.Name, prefixMaster); err != nil {
						s.logger.Log("error", microerror.MaskAny(err))
						return
					}

					// Run workers
					if err := s.runMachines(cluster.Spec, cluster.Name, prefixWorker); err != nil {
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

func (s *Service) runMachines(spec awstpr.Spec, clusterName string, prefix string) error {
	var (
		machines        []node.Node
		awsMachines     []tpraws.Node
		cloudconfigPath string
	)

	switch prefix {
	case prefixMaster:
		machines = spec.Masters
		awsMachines = spec.Aws.Masters
		cloudconfigPath = s.cloudconfigMasterPath
	case prefixWorker:
		machines = spec.Workers
		awsMachines = spec.Aws.Workers
		cloudconfigPath = s.cloudconfigWorkerPath
	default:
		return microerror.MaskAny(fmt.Errorf("invalid prefix %q", prefix))
	}

	if len(machines) != len(awsMachines) {
		return microerror.MaskAny(fmt.Errorf("mismatched number of %q machines in the 'spec' and 'aws' sections: %d != %d",
			prefix,
			len(machines),
			len(awsMachines)))
	}

	for no, machine := range machines {
		name := fmt.Sprintf("%s-%d", prefix, no)
		m := awsNode{
			Node:    machine,
			AwsInfo: awsMachines[no],
		}
		if err := s.runMachine(m, spec, clusterName, cloudconfigPath, name); err != nil {
			return microerror.MaskAny(err)
		}
	}
	return nil
}

const (
	roleName                 = "EC2-K8S-Role"
	policyName               = "EC2-K8S-Policy"
	profileName              = "EC2-DecryptTLSCerts"
	assumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": {
			"Effect": "Allow",
			"Principal": {
				"Service": "ec2.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}
	}`
	policyDocumentTempl = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "kms:Decrypt",
				"Resource": %q
			}
		]
	}`
)

func (s *Service) encodeTLSAssets(awsSession *session.Session, kmsKeyArn string) (*CompactTLSAssets, error) {
	rawTLS, err := readRawTLSAssets(s.certsDir)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	policyDocument := fmt.Sprintf(policyDocumentTempl, kmsKeyArn)

	svc := iam.New(awsSession)

	if _, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDocument),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				s.logger.Log("info", fmt.Sprintf("role '%s' already exists, reusing", roleName))
			default:
				return nil, microerror.MaskAny(err)
			}
		}
	}

	if _, err := svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(policyName),
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(policyDocument),
	}); err != nil {
		return nil, microerror.MaskAny(err)
	}

	_, err = svc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	})
	switch {
	case err != nil:
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				s.logger.Log("info", fmt.Sprintf("instance profile '%s' already exists, reusing", roleName))
			default:
				return nil, microerror.MaskAny(err)
			}
		}
	default:
		if _, err := svc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(profileName),
			RoleName:            aws.String(roleName),
		}); err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	encTLS, err := rawTLS.encrypt(awsSession, kmsKeyArn)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	compTLS, err := encTLS.compact()
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return compTLS, nil
}

func (s *Service) runMachine(machine awsNode, spec awstpr.Spec, clusterName, cloudconfigPath, name string) error {
	s.awsConfig.Region = spec.Aws.Region
	awsSession, ec2Client := awsutil.NewClient(s.awsConfig)

	instances, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
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

	tlsAssets, err := s.encodeTLSAssets(awsSession, spec.Aws.KMSKeyArn)
	if err != nil {
		return microerror.MaskAny(err)
	}

	params := cloudconfigTemplateParams{
		Spec:      spec,
		Node:      machine,
		TLSAssets: *tlsAssets,
	}

	cloudconfig, err := newCloudConfig(cloudconfigPath, params)
	if err != nil {
		return microerror.MaskAny(err)
	}
	if err := cloudconfig.executeTemplate(); err != nil {
		return microerror.MaskAny(err)
	}
	cloudconfigBase64 := cloudconfig.base64()

	// add instance profile to reservation
	reservation, err := ec2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(machine.AwsInfo.ImageID),
		InstanceType: aws.String(machine.AwsInfo.InstanceType),
		MinCount:     aws.Int64(int64(1)),
		MaxCount:     aws.Int64(int64(1)),
		UserData:     &cloudconfigBase64,
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(profileName),
		},
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("instance '%s' reserved", name))

	if _, err := ec2Client.CreateTags(&ec2.CreateTagsInput{
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
