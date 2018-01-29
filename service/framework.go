package service

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeytpr"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfigv2"
	"github.com/giantswarm/aws-operator/service/cloudconfigv3"
	"github.com/giantswarm/aws-operator/service/cloudconfigv4"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2/adapter"
	"github.com/giantswarm/aws-operator/service/resource/endpointsv2"
	"github.com/giantswarm/aws-operator/service/resource/kmskeyv2"
	"github.com/giantswarm/aws-operator/service/resource/legacyv2"
	"github.com/giantswarm/aws-operator/service/resource/namespacev2"
	"github.com/giantswarm/aws-operator/service/resource/s3bucketv2"
	"github.com/giantswarm/aws-operator/service/resource/s3objectv2"
	"github.com/giantswarm/aws-operator/service/resource/servicev2"
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

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	clientSet, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
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

	versionedResources, err := NewVersionedResources(config, k8sClient, awsConfig, awsHostConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = clientSet.ProviderV1alpha1().AWSConfigs("")

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

func NewVersionedResources(config Config, k8sClient kubernetes.Interface, awsConfig awsclient.Config, awsHostConfig awsclient.Config) (map[string][]framework.Resource, error) {
	var err error

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

	var certWatcher *legacy.Service
	{
		certConfig := legacy.DefaultServiceConfig()
		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = legacy.NewService(certConfig)
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

	// ccServiceV2 is used by the legacyv2 resource.
	var ccServiceV2 *cloudconfigv2.CloudConfig
	{
		ccServiceConfig := cloudconfigv2.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV2, err = cloudconfigv2.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// ccServiceV3 is used by the s3objectv2 resource for s3BucketObjectResourceV1.
	var ccServiceV3 *cloudconfigv3.CloudConfig
	{
		ccServiceConfig := cloudconfigv3.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV3, err = cloudconfigv3.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// ccServiceV4 is used by the s3objectv2 resource for s3BucketObjectResourceV2.
	var ccServiceV4 *cloudconfigv4.CloudConfig
	{
		ccServiceConfig := cloudconfigv4.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV4, err = cloudconfigv4.New(ccServiceConfig)
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

	var s3BucketObjectResourceV1 framework.Resource
	{
		s3BucketObjectConfig := s3objectv2.DefaultConfig()
		s3BucketObjectConfig.AwsService = awsService
		s3BucketObjectConfig.Clients.S3 = awsClients.S3
		s3BucketObjectConfig.Clients.KMS = awsClients.KMS
		s3BucketObjectConfig.CloudConfig = ccServiceV3
		s3BucketObjectConfig.CertWatcher = certWatcher
		s3BucketObjectConfig.Logger = config.Logger
		s3BucketObjectConfig.RandomKeyWatcher = keyWatcher

		s3BucketObjectResourceV1, err = s3objectv2.New(s3BucketObjectConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketObjectResourceV2 framework.Resource
	{
		c := s3objectv2.DefaultConfig()
		c.AwsService = awsService
		c.Clients.S3 = awsClients.S3
		c.Clients.KMS = awsClients.KMS
		c.CloudConfig = ccServiceV4
		c.CertWatcher = certWatcher
		c.Logger = config.Logger
		c.RandomKeyWatcher = keyWatcher

		s3BucketObjectResourceV2, err = s3objectv2.New(c)
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

	var serviceResource framework.Resource
	{
		serviceConfig := servicev2.DefaultConfig()

		serviceConfig.K8sClient = k8sClient
		serviceConfig.Logger = config.Logger

		serviceResource, err = servicev2.New(serviceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var endpointsResource framework.Resource
	{
		endpointsConfig := endpointsv2.DefaultConfig()

		endpointsConfig.Clients.EC2 = awsClients.EC2
		endpointsConfig.K8sClient = k8sClient
		endpointsConfig.Logger = config.Logger

		endpointsResource, err = endpointsv2.New(endpointsConfig)
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
	var resourcesV1 []framework.Resource
	{
		resourcesV1 = []framework.Resource{
			kmsKeyResource,
			s3BucketResource,
			s3BucketObjectResourceV1,
			cloudformationResource,
			namespaceResource,
			serviceResource,
			endpointsResource,
		}

		resourcesV1, err = metricsresource.Wrap(resourcesV1, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We create the list of resources and wrap each resource around some common
	// resources like metrics and retry resources.
	var resourcesV2 []framework.Resource
	{
		resourcesV2 = []framework.Resource{
			kmsKeyResource,
			s3BucketResource,
			s3BucketObjectResourceV2,
			cloudformationResource,
			namespaceResource,
			serviceResource,
			endpointsResource,
		}

		resourcesV2, err = metricsresource.Wrap(resourcesV2, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We provide a map of resource lists keyed by the version bundle version
	// to the resource router.
	versionedResources := map[string][]framework.Resource{
		// Clusters without a version use the legacy resource.
		"":      legacyResources,
		"0.1.0": legacyResources,
		"0.2.0": resourcesV1,
		"1.0.0": legacyResources,
		"2.0.0": resourcesV1,
		"2.1.0": resourcesV2,
	}

	return versionedResources, nil
}
