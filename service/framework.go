package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/clustertpr/spec/kubernetes/ssh"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/client/k8sextclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/giantswarm/randomkeytpr"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfigv1"
	"github.com/giantswarm/aws-operator/service/cloudconfigv2"
	"github.com/giantswarm/aws-operator/service/cloudconfigv3"
	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2/adapter"
	"github.com/giantswarm/aws-operator/service/resource/kmskeyv2"
	"github.com/giantswarm/aws-operator/service/resource/legacyv1"
	"github.com/giantswarm/aws-operator/service/resource/legacyv2"
	"github.com/giantswarm/aws-operator/service/resource/namespacev1"
	"github.com/giantswarm/aws-operator/service/resource/namespacev2"
	"github.com/giantswarm/aws-operator/service/resource/s3bucketv1"
	"github.com/giantswarm/aws-operator/service/resource/s3bucketv2"
	"github.com/giantswarm/aws-operator/service/resource/s3objectv1"
	"github.com/giantswarm/aws-operator/service/resource/s3objectv2"
)

const (
	ResourceRetries  uint64 = 3
	awsCloudProvider        = "aws"
)

const (
	AWSConfigCleanupFinalizer = "aws-operator.giantswarm.io/custom-object-cleanup"
)

func newCRDFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var k8sExtClient apiextensionsclient.Interface
	{
		c := k8sextclient.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sExtClient, err = k8sextclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = k8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient kubernetes.Interface
	{
		c := k8sclient.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultServiceConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keyWatcher *randomkeytpr.Service
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = k8sClient
		keyConfig.Logger = config.Logger
		keyWatcher, err = randomkeytpr.NewService(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsConfig awsclient.Config
	{
		awsConfig = awsclient.Config{
			AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
			AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
			SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
			Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
		}
	}

	var awsHostConfig awsclient.Config
	{
		accessKeyID := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID)
		accessKeySecret := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret)
		sessionToken := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session)

		if accessKeyID == "" && accessKeySecret == "" {
			config.Logger.Log("debug", "no host cluster account credentials supplied, assuming guest and host uses same account")
			awsHostConfig = awsConfig
		} else {
			config.Logger.Log("debug", "host cluster account credentials supplied, using separate accounts for host and guest clusters")
			awsHostConfig = awsclient.Config{
				AccessKeyID:     accessKeyID,
				AccessKeySecret: accessKeySecret,
				SessionToken:    sessionToken,
			}
		}
	}

	awsClients := awsclient.NewClients(awsConfig)
	var awsService *awsservice.Service
	{
		awsConfig := awsservice.DefaultConfig()
		awsConfig.Clients.IAM = awsClients.IAM
		awsConfig.Clients.KMS = awsClients.KMS
		awsConfig.Logger = config.Logger

		awsService, err = awsservice.New(awsConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	awsHostClients := awsclient.NewClients(awsHostConfig)

	// ccServicev2 is used by the legacyv2 resource.
	var ccServiceV2 *cloudconfigv2.CloudConfig
	{
		ccServiceConfig := cloudconfigv2.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV2, err = cloudconfigv2.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// ccServicev3 is used by the s3objectv2 resource.
	var ccServiceV3 *cloudconfigv3.CloudConfig
	{
		ccServiceConfig := cloudconfigv3.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV3, err = cloudconfigv3.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	installationName := config.Viper.GetString(config.Flag.Service.Installation.Name)

	var legacyResource framework.Resource
	{
		legacyConfig := legacyv2.DefaultConfig()
		legacyConfig.AwsConfig = awsConfig
		legacyConfig.AwsHostConfig = awsHostConfig
		legacyConfig.CertWatcher = certWatcher
		legacyConfig.CloudConfig = ccServiceV2
		legacyConfig.InstallationName = installationName
		legacyConfig.K8sClient = k8sClient
		legacyConfig.KeyWatcher = keyWatcher
		legacyConfig.Logger = config.Logger
		legacyConfig.PubKeyFile = config.Viper.GetString(config.Flag.Service.AWS.PubKeyFile)

		legacyResource, err = legacyv2.New(legacyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudformationResource framework.Resource
	{
		cloudformationConfig := cloudformationv2.DefaultConfig()

		cloudformationConfig.Clients = &adapter.Clients{}
		cloudformationConfig.Clients.EC2 = awsClients.EC2
		cloudformationConfig.Clients.CloudFormation = awsClients.CloudFormation
		cloudformationConfig.Clients.IAM = awsClients.IAM
		cloudformationConfig.Clients.KMS = awsClients.KMS
		cloudformationConfig.Clients.ELB = awsClients.ELB

		cloudformationConfig.HostClients = &adapter.Clients{}
		cloudformationConfig.HostClients.EC2 = awsHostClients.EC2
		cloudformationConfig.HostClients.IAM = awsHostClients.IAM
		cloudformationConfig.HostClients.CloudFormation = awsHostClients.CloudFormation

		cloudformationConfig.Logger = config.Logger

		cloudformationConfig.InstallationName = installationName

		cloudformationResource, err = cloudformationv2.New(cloudformationConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kmsKeyResource framework.Resource
	{
		kmsKeyConfig := kmskeyv2.DefaultConfig()
		kmsKeyConfig.Clients.KMS = awsClients.KMS
		kmsKeyConfig.Logger = config.Logger

		kmsKeyResource, err = kmskeyv2.New(kmsKeyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketResource framework.Resource
	{
		s3BucketConfig := s3bucketv2.DefaultConfig()
		s3BucketConfig.AwsService = awsService
		s3BucketConfig.Clients.S3 = awsClients.S3
		s3BucketConfig.Logger = config.Logger

		s3BucketResource, err = s3bucketv2.New(s3BucketConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketObjectResource framework.Resource
	{
		s3BucketObjectConfig := s3objectv2.DefaultConfig()
		s3BucketObjectConfig.AwsService = awsService
		s3BucketObjectConfig.Clients.S3 = awsClients.S3
		s3BucketObjectConfig.Clients.KMS = awsClients.KMS
		s3BucketObjectConfig.CloudConfig = ccServiceV3
		s3BucketObjectConfig.CertWatcher = certWatcher
		s3BucketObjectConfig.Logger = config.Logger
		s3BucketObjectConfig.RandomKeyWatcher = keyWatcher

		s3BucketObjectResource, err = s3objectv2.New(s3BucketObjectConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		namespaceConfig := namespacev2.DefaultConfig()

		namespaceConfig.K8sClient = k8sClient
		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespacev2.New(namespaceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// Metrics config for wrapping resources.
	metricsWrapConfig := metricsresource.DefaultWrapConfig()
	metricsWrapConfig.Name = config.Name

	// Existing clusters are only processed by the legacy resource. We wrap it
	// with the metrics resource for monitoring.
	var legacyResources []framework.Resource
	{
		legacyResources = []framework.Resource{
			legacyResource,
		}

		legacyResources, err = metricsresource.Wrap(legacyResources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We create the list of resources and wrap each resource around some common
	// resources like metrics and retry resources.
	// TODO Remove the legacy resource once all resources are migrated to
	// Cloud Formation.
	var resources []framework.Resource
	{
		resources = []framework.Resource{
			namespaceResource,
			kmsKeyResource,
			s3BucketResource,
			s3BucketObjectResource,
			legacyResource,
			cloudformationResource,
		}

		// Disable retry wrapper due to problems with the legacy resource.
		//
		// NOTE that the retry resources wrap the underlying resources first. The
		// wrapped resources are then wrapped around the metrics resource. That way
		// the metrics also consider execution times and execution attempts including
		// retries.
		/*
			retryWrapConfig := retryresource.DefaultWrapConfig()
			retryWrapConfig.Logger = config.Logger
			cloudFormationResources, err = retryresource.Wrap(cloudFormationResources, retryWrapConfig)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		*/

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We provide a map of resource lists keyed by the version bundle version
	// to the resource router.
	versionedResources := map[string][]framework.Resource{
		keyv2.LegacyVersion:         legacyResources,
		keyv2.CloudFormationVersion: resources,
		"1.0.0":                     legacyResources,
	}

	var clientSet *versioned.Clientset
	{
		var c *rest.Config

		if config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster) {
			config.Logger.Log("debug", "creating in-cluster config")

			c, err = rest.InClusterConfig()
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			config.Logger.Log("debug", "creating out-cluster config")

			c = &rest.Config{
				Host: config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
					CertFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
					KeyFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
				},
			}
		}

		clientSet, err = versioned.NewForConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// TODO remove after migration.
	migrateTPRsToCRDs(config.Logger, clientSet)

	var newWatcherFactory informer.WatcherFactory
	{
		newWatcherFactory = func() (watch.Interface, error) {
			watcher, err := clientSet.ProviderV1alpha1().AWSConfigs("").Watch(apismetav1.ListOptions{})
			if err != nil {
				return nil, microerror.Mask(err)
			}

			return watcher, nil
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.WatcherFactory = newWatcherFactory

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	var crdFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.CRD = v1alpha1.NewAWSConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.InitCtxFunc = initCtxFunc
		c.Logger = config.Logger
		c.ResourceRouter = NewResourceRouter(versionedResources)

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}

func newCustomObjectFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()
		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.Logger = config.Logger
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultServiceConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = certificatetpr.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keyWatcher *randomkeytpr.Service
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = k8sClient
		keyConfig.Logger = config.Logger
		keyWatcher, err = randomkeytpr.NewService(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsConfig awsclient.Config
	{
		awsConfig = awsclient.Config{
			AccessKeyID:     config.Viper.GetString(config.Flag.Service.AWS.AccessKey.ID),
			AccessKeySecret: config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Secret),
			SessionToken:    config.Viper.GetString(config.Flag.Service.AWS.AccessKey.Session),
			Region:          config.Viper.GetString(config.Flag.Service.AWS.Region),
		}
	}

	var awsHostConfig awsclient.Config
	{
		accessKeyID := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.ID)
		accessKeySecret := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Secret)
		sessionToken := config.Viper.GetString(config.Flag.Service.AWS.HostAccessKey.Session)

		if accessKeyID == "" && accessKeySecret == "" {
			config.Logger.Log("debug", "no host cluster account credentials supplied, assuming guest and host uses same account")
			awsHostConfig = awsConfig
		} else {
			config.Logger.Log("debug", "host cluster account credentials supplied, using separate accounts for host and guest clusters")
			awsHostConfig = awsclient.Config{
				AccessKeyID:     accessKeyID,
				AccessKeySecret: accessKeySecret,
				SessionToken:    sessionToken,
			}
		}
	}

	awsClients := awsclient.NewClients(awsConfig)
	var awsService *awsservice.Service
	{
		awsConfig := awsservice.DefaultConfig()
		awsConfig.Clients.IAM = awsClients.IAM
		awsConfig.Clients.KMS = awsClients.KMS
		awsConfig.Logger = config.Logger

		awsService, err = awsservice.New(awsConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ccService *cloudconfigv1.CloudConfig
	{
		ccServiceConfig := cloudconfigv1.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccService, err = cloudconfigv1.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	installationName := config.Viper.GetString(config.Flag.Service.Installation.Name)

	var legacyResource framework.Resource
	{
		legacyConfig := legacyv1.DefaultConfig()
		legacyConfig.AwsConfig = awsConfig
		legacyConfig.AwsHostConfig = awsHostConfig
		legacyConfig.CertWatcher = certWatcher
		legacyConfig.CloudConfig = ccService
		legacyConfig.InstallationName = installationName
		legacyConfig.K8sClient = k8sClient
		legacyConfig.KeyWatcher = keyWatcher
		legacyConfig.Logger = config.Logger
		legacyConfig.PubKeyFile = config.Viper.GetString(config.Flag.Service.AWS.PubKeyFile)

		legacyResource, err = legacyv1.New(legacyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketResource framework.Resource
	{
		s3BucketConfig := s3bucketv1.DefaultConfig()
		s3BucketConfig.AwsService = awsService
		s3BucketConfig.Clients.S3 = awsClients.S3
		s3BucketConfig.Logger = config.Logger

		s3BucketResource, err = s3bucketv1.New(s3BucketConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketObjectResource framework.Resource
	{
		s3BucketObjectConfig := s3objectv1.DefaultConfig()
		s3BucketObjectConfig.AwsService = awsService
		s3BucketObjectConfig.Clients.S3 = awsClients.S3
		s3BucketObjectConfig.Clients.KMS = awsClients.KMS
		s3BucketObjectConfig.CloudConfig = ccService
		s3BucketObjectConfig.CertWatcher = certWatcher
		s3BucketObjectConfig.Logger = config.Logger

		s3BucketObjectResource, err = s3objectv1.New(s3BucketObjectConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		namespaceConfig := namespacev1.DefaultConfig()

		namespaceConfig.K8sClient = k8sClient
		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespacev1.New(namespaceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// Metrics config for wrapping resources.
	metricsWrapConfig := metricsresource.DefaultWrapConfig()
	metricsWrapConfig.Name = config.Name

	// Existing clusters are only processed by the legacy resource. We wrap it
	// with the metrics resource for monitoring.
	var legacyResources []framework.Resource
	{
		legacyResources = []framework.Resource{
			legacyResource,
		}

		legacyResources, err = metricsresource.Wrap(legacyResources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We create the list of resources and wrap each resource around some common
	// resources like metrics and retry resources.
	// TODO Remove the legacy resource once all resources are migrated to
	// Cloud Formation.
	var resources []framework.Resource
	{
		resources = []framework.Resource{
			namespaceResource,
			legacyResource,
			s3BucketResource,
			s3BucketObjectResource,
		}

		// Disable retry wrapper due to problems with the legacy resource.
		//
		// NOTE that the retry resources wrap the underlying resources first. The
		// wrapped resources are then wrapped around the metrics resource. That way
		// the metrics also consider execution times and execution attempts including
		// retries.
		/*
			retryWrapConfig := retryresource.DefaultWrapConfig()
			retryWrapConfig.Logger = config.Logger
			cloudFormationResources, err = retryresource.Wrap(cloudFormationResources, retryWrapConfig)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		*/

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We provide a map of resource lists keyed by the version bundle version
	// to the resource router.
	versionedResources := map[string][]framework.Resource{
		keyv1.LegacyVersion:         legacyResources,
		keyv1.CloudFormationVersion: resources,
		"1.0.0":                     legacyResources,
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	var newTPR *tpr.TPR
	{
		c := tpr.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		c.Description = awstpr.Description
		c.Name = awstpr.Name
		c.Version = awstpr.VersionV1

		newTPR, err = tpr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newWatcherFactory informer.WatcherFactory
	{
		zeroObjectFactory := &informer.ZeroObjectFactoryFuncs{
			NewObjectFunc:     func() runtime.Object { return &awstpr.CustomObject{} },
			NewObjectListFunc: func() runtime.Object { return &awstpr.List{} },
		}
		newWatcherFactory = informer.NewWatcherFactory(k8sClient.Discovery().RESTClient(), newTPR.WatchEndpoint(""), zeroObjectFactory)
	}

	var newInformer *informer.Informer
	{
		informerConfig := informer.DefaultConfig()

		informerConfig.WatcherFactory = newWatcherFactory
		informerConfig.RateWait = time.Second * 10
		informerConfig.ResyncPeriod = time.Minute * 5

		newInformer, err = informer.New(informerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var customObjectFramework *framework.Framework
	{
		frameworkConfig := framework.DefaultConfig()

		frameworkConfig.InitCtxFunc = initCtxFunc
		frameworkConfig.Logger = config.Logger
		frameworkConfig.ResourceRouter = NewTPRResourceRouter(versionedResources)
		frameworkConfig.Informer = newInformer
		frameworkConfig.TPR = newTPR

		customObjectFramework, err = framework.New(frameworkConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return customObjectFramework, nil
}

func migrateTPRsToCRDs(logger micrologger.Logger, clientSet *versioned.Clientset) {
	logger.Log("debug", "start TPR migration")

	var err error

	// List all TPOs.
	var b []byte
	{
		e := "/apis/cluster.giantswarm.io/v1/namespaces/default/awses"
		b, err = clientSet.Discovery().RESTClient().Get().AbsPath(e).DoRaw()
		if err != nil {
			logger.Log("error", fmt.Sprintf("%#v", err))
			return
		}

		fmt.Printf("\n")
		fmt.Printf("b start\n")
		fmt.Printf("%s\n", b)
		fmt.Printf("b end\n")
		fmt.Printf("\n")
	}

	// Convert bytes into structure.
	var v *awstpr.List
	{
		v = &awstpr.List{}
		if err := json.Unmarshal(b, v); err != nil {
			logger.Log("error", fmt.Sprintf("%#v", err))
			return
		}

		fmt.Printf("\n")
		fmt.Printf("v start\n")
		fmt.Printf("%#v\n", v)
		fmt.Printf("v end\n")
		fmt.Printf("\n")
	}

	// Iterate over all TPOs.
	for _, tpo := range v.Items {
		// Compute CRO using TPO.
		var cro *v1alpha1.AWSConfig
		{
			cro = &v1alpha1.AWSConfig{}

			cro.TypeMeta.APIVersion = "provider.giantswarm.io"
			cro.TypeMeta.Kind = "AWSConfig"
			cro.ObjectMeta.Name = tpo.Name
			//cro.ObjectMeta.Finalizers = []string{
			//	AWSConfigCleanupFinalizer,
			//}
			cro.Spec.AWS.API.ELB.IdleTimeoutSeconds = tpo.Spec.AWS.ELB.IdleTimeoutSeconds.API
			cro.Spec.AWS.API.HostedZones = tpo.Spec.AWS.HostedZones.API
			cro.Spec.AWS.AZ = tpo.Spec.AWS.AZ
			cro.Spec.AWS.Etcd.ELB.IdleTimeoutSeconds = tpo.Spec.AWS.ELB.IdleTimeoutSeconds.Etcd
			cro.Spec.AWS.Etcd.HostedZones = tpo.Spec.AWS.HostedZones.Etcd
			cro.Spec.AWS.Ingress.ELB.IdleTimeoutSeconds = tpo.Spec.AWS.ELB.IdleTimeoutSeconds.Ingress
			cro.Spec.AWS.Ingress.HostedZones = tpo.Spec.AWS.HostedZones.Ingress
			cro.Spec.AWS.Masters = toAWSMasters(tpo.Spec.AWS.Masters)
			cro.Spec.AWS.Region = tpo.Spec.AWS.Region
			cro.Spec.AWS.VPC.CIDR = tpo.Spec.AWS.VPC.CIDR
			cro.Spec.AWS.VPC.PeerID = tpo.Spec.AWS.VPC.PeerID
			cro.Spec.AWS.VPC.PrivateSubnetCIDR = tpo.Spec.AWS.VPC.PrivateSubnetCIDR
			cro.Spec.AWS.VPC.PublicSubnetCIDR = tpo.Spec.AWS.VPC.PublicSubnetCIDR
			cro.Spec.AWS.VPC.RouteTableNames = tpo.Spec.AWS.VPC.RouteTableNames
			cro.Spec.AWS.Workers = toAWSWorkers(tpo.Spec.AWS.Workers)
			cro.Spec.Cluster.Calico.CIDR = tpo.Spec.Cluster.Calico.CIDR
			cro.Spec.Cluster.Calico.Domain = tpo.Spec.Cluster.Calico.Domain
			cro.Spec.Cluster.Calico.MTU = tpo.Spec.Cluster.Calico.MTU
			cro.Spec.Cluster.Calico.Subnet = tpo.Spec.Cluster.Calico.Subnet
			cro.Spec.Cluster.Customer.ID = tpo.Spec.Cluster.Customer.ID
			cro.Spec.Cluster.Docker.Daemon.CIDR = tpo.Spec.Cluster.Docker.Daemon.CIDR
			cro.Spec.Cluster.Docker.Daemon.ExtraArgs = tpo.Spec.Cluster.Docker.Daemon.ExtraArgs
			cro.Spec.Cluster.Etcd.AltNames = tpo.Spec.Cluster.Etcd.AltNames
			cro.Spec.Cluster.Etcd.Domain = tpo.Spec.Cluster.Etcd.Domain
			cro.Spec.Cluster.Etcd.Port = tpo.Spec.Cluster.Etcd.Port
			cro.Spec.Cluster.Etcd.Prefix = tpo.Spec.Cluster.Etcd.Prefix
			cro.Spec.Cluster.ID = tpo.Spec.Cluster.Cluster.ID
			cro.Spec.Cluster.Kubernetes.API.AltNames = tpo.Spec.Cluster.Kubernetes.API.AltNames
			cro.Spec.Cluster.Kubernetes.API.ClusterIPRange = tpo.Spec.Cluster.Kubernetes.API.ClusterIPRange
			cro.Spec.Cluster.Kubernetes.API.Domain = tpo.Spec.Cluster.Kubernetes.API.Domain
			cro.Spec.Cluster.Kubernetes.API.InsecurePort = tpo.Spec.Cluster.Kubernetes.API.InsecurePort
			cro.Spec.Cluster.Kubernetes.API.IP = tpo.Spec.Cluster.Kubernetes.API.IP
			cro.Spec.Cluster.Kubernetes.API.SecurePort = tpo.Spec.Cluster.Kubernetes.API.SecurePort
			cro.Spec.Cluster.Kubernetes.CloudProvider = tpo.Spec.Cluster.Kubernetes.CloudProvider
			cro.Spec.Cluster.Kubernetes.DNS.IP = tpo.Spec.Cluster.Kubernetes.DNS.IP
			cro.Spec.Cluster.Kubernetes.Domain = tpo.Spec.Cluster.Kubernetes.Domain
			cro.Spec.Cluster.Kubernetes.Hyperkube.Docker.Image = tpo.Spec.Cluster.Kubernetes.Hyperkube.Docker.Image
			cro.Spec.Cluster.Kubernetes.IngressController.Docker.Image = tpo.Spec.Cluster.Kubernetes.IngressController.Docker.Image
			cro.Spec.Cluster.Kubernetes.IngressController.Domain = tpo.Spec.Cluster.Kubernetes.IngressController.Domain
			cro.Spec.Cluster.Kubernetes.IngressController.InsecurePort = tpo.Spec.Cluster.Kubernetes.IngressController.InsecurePort
			cro.Spec.Cluster.Kubernetes.IngressController.SecurePort = tpo.Spec.Cluster.Kubernetes.IngressController.SecurePort
			cro.Spec.Cluster.Kubernetes.IngressController.WildcardDomain = tpo.Spec.Cluster.Kubernetes.IngressController.WildcardDomain
			cro.Spec.Cluster.Kubernetes.Kubelet.AltNames = tpo.Spec.Cluster.Kubernetes.Kubelet.AltNames
			cro.Spec.Cluster.Kubernetes.Kubelet.Domain = tpo.Spec.Cluster.Kubernetes.Kubelet.Domain
			cro.Spec.Cluster.Kubernetes.Kubelet.Labels = tpo.Spec.Cluster.Kubernetes.Kubelet.Labels
			cro.Spec.Cluster.Kubernetes.Kubelet.Port = tpo.Spec.Cluster.Kubernetes.Kubelet.Port
			cro.Spec.Cluster.Kubernetes.NetworkSetup.Docker.Image = tpo.Spec.Cluster.Kubernetes.NetworkSetup.Docker.Image
			cro.Spec.Cluster.Kubernetes.SSH.UserList = toUserList(tpo.Spec.Cluster.Kubernetes.SSH.UserList)
			cro.Spec.Cluster.Masters = toClusterMasters(tpo.Spec.Cluster.Masters)
			cro.Spec.Cluster.Vault.Address = tpo.Spec.Cluster.Vault.Address
			cro.Spec.Cluster.Vault.Token = tpo.Spec.Cluster.Vault.Token
			cro.Spec.Cluster.Version = tpo.Spec.Cluster.Version
			cro.Spec.Cluster.Workers = toClusterWorkers(tpo.Spec.Cluster.Workers)
			cro.Spec.VersionBundle.Version = tpo.Spec.VersionBundle.Version

			fmt.Printf("\n")
			fmt.Printf("cro start\n")
			fmt.Printf("%#v\n", cro)
			fmt.Printf("cro end\n")
			fmt.Printf("\n")
		}

		// Create CRO in Kubernetes API.
		{
			_, err := clientSet.ProviderV1alpha1().AWSConfigs(tpo.Namespace).Get(cro.Name, apismetav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				_, err := clientSet.ProviderV1alpha1().AWSConfigs(tpo.Namespace).Create(cro)
				if err != nil {
					logger.Log("error", fmt.Sprintf("%#v", err))
					return
				}
			} else if err != nil {
				logger.Log("error", fmt.Sprintf("%#v", err))
				return
			}
		}
	}

	// update existing CROs with empty cloud provider
	cros, err := clientSet.
		ProviderV1alpha1().
		AWSConfigs("default").
		List(apismetav1.ListOptions{})
	if err != nil {
		logger.Log("error", fmt.Sprintf("%#v", err))
		return
	}
	for _, cro := range cros.Items {
		if cro.Spec.Cluster.Kubernetes.CloudProvider != "" {
			continue
		}
		// CRO existed with empty CloudProvider, refresh
		type PatchSpec struct {
			Op    string `json:"op"`
			Path  string `json:"path"`
			Value string `json:"value"`
		}
		patch := make([]PatchSpec, 1)
		patch[0].Op = "replace"
		patch[0].Path = "/spec/cluster/kubernetes/cloudProvider"
		patch[0].Value = awsCloudProvider
		patchBytes, err := json.Marshal(patch)
		if err != nil {
			logger.Log("error", fmt.Sprintf("%#v", err))
		}
		_, err = clientSet.
			ProviderV1alpha1().
			AWSConfigs("default").
			Patch(cro.Name, types.JSONPatchType, patchBytes)
		if err != nil {
			logger.Log("error", fmt.Sprintf("%#v", err))
		}
	}
	logger.Log("debug", "end TPR migration")
}

func toClusterMasters(masters []spec.Node) []v1alpha1.ClusterNode {
	var newList []v1alpha1.ClusterNode

	for _, m := range masters {
		n := v1alpha1.ClusterNode{
			ID: m.ID,
		}

		newList = append(newList, n)
	}

	return newList
}

func toClusterWorkers(workers []spec.Node) []v1alpha1.ClusterNode {
	var newList []v1alpha1.ClusterNode

	for _, w := range workers {
		n := v1alpha1.ClusterNode{
			ID: w.ID,
		}

		newList = append(newList, n)
	}

	return newList
}

func toAWSMasters(masters []aws.Node) []v1alpha1.AWSConfigSpecAWSNode {
	var newList []v1alpha1.AWSConfigSpecAWSNode

	for _, m := range masters {
		n := v1alpha1.AWSConfigSpecAWSNode{
			ImageID:      m.ImageID,
			InstanceType: m.InstanceType,
		}
		newList = append(newList, n)
	}

	return newList
}

func toAWSWorkers(workers []aws.Node) []v1alpha1.AWSConfigSpecAWSNode {
	var newList []v1alpha1.AWSConfigSpecAWSNode

	for _, w := range workers {
		n := v1alpha1.AWSConfigSpecAWSNode{
			ImageID:      w.ImageID,
			InstanceType: w.InstanceType,
		}
		newList = append(newList, n)
	}

	return newList
}

func toUserList(userList []ssh.User) []v1alpha1.ClusterKubernetesSSHUser {
	var newList []v1alpha1.ClusterKubernetesSSHUser

	for _, user := range userList {
		u := v1alpha1.ClusterKubernetesSSHUser{
			Name:      user.Name,
			PublicKey: user.PublicKey,
		}

		newList = append(newList, u)
	}

	return newList
}
