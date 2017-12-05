package service

import (
	"context"
	"time"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfig"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv1"
	"github.com/giantswarm/aws-operator/service/resource/legacyv1"
	"github.com/giantswarm/aws-operator/service/resource/namespacev1"
	"github.com/giantswarm/aws-operator/service/resource/s3bucketv1"
	"github.com/giantswarm/aws-operator/service/resource/s3objectv1"
)

const (
	ResourceRetries uint64 = 3
)

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

	var ccService *cloudconfig.CloudConfig
	{
		ccServiceConfig := cloudconfig.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccService, err = cloudconfig.New(ccServiceConfig)
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

	var cloudformationResource framework.Resource
	{
		cloudformationConfig := cloudformationv1.DefaultConfig()

		cloudformationConfig.Clients.EC2 = awsClients.EC2
		cloudformationConfig.Clients.CloudFormation = awsClients.CloudFormation
		cloudformationConfig.Clients.IAM = awsClients.IAM
		cloudformationConfig.Logger = config.Logger

		cloudformationResource, err = cloudformationv1.New(cloudformationConfig)
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
		key.LegacyVersion:         legacyResources,
		key.CloudFormationVersion: resources,
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
		frameworkConfig.ResourceRouter = NewResourceRouter(versionedResources)
		frameworkConfig.Informer = newInformer
		frameworkConfig.TPR = newTPR

		customObjectFramework, err = framework.New(frameworkConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return customObjectFramework, nil
}
