package service

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeytpr"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	cloudconfigv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/cloudconfig"
	cloudformationv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/cloudformation/adapter"
	endpointsv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/endpoints"
	kmskeyv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/kmskey"
	legacyv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/legacy"
	namespacev2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/namespace"
	s3bucketv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/s3bucket"
	s3objectv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/s3object"
	servicev2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/service"
	cloudconfigv3 "github.com/giantswarm/aws-operator/service/awsconfig/v3/cloudconfig"
	s3objectv3 "github.com/giantswarm/aws-operator/service/awsconfig/v3/resource/s3object"
	cloudconfigv4 "github.com/giantswarm/aws-operator/service/awsconfig/v4/cloudconfig"
)

const (
	ResourceRetries  = 3
	awsCloudProvider = "aws"
)

const (
	AWSConfigCleanupFinalizer = "aws-operator.giantswarm.io/custom-object-cleanup"
)

type FrameworkConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestAWSConfig   FrameworkConfigAWSConfig
	HostAWSConfig    FrameworkConfigAWSConfig
	InstallationName string
	// Name is the name of the project.
	Name       string
	PubKeyFile string
}

type FrameworkConfigAWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	SessionToken    string
}

func NewFramework(config FrameworkConfig) (*framework.Framework, error) {
	var err error

	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.GuestAWSConfig.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.AccessKeyID must not be empty")
	}
	if config.GuestAWSConfig.AccessKeySecret == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.AccessKeySecret must not be empty")
	}
	if config.GuestAWSConfig.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.Region must not be empty")
	}
	if config.GuestAWSConfig.SessionToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.SessionToken must not be empty")
	}
	if config.HostAWSConfig.AccessKeyID == "" && config.HostAWSConfig.AccessKeySecret == "" {
		config.Logger.Log("debug", "no host cluster account credentials supplied, assuming guest and host uses same account")
		config.HostAWSConfig = config.GuestAWSConfig
	} else {
		if config.HostAWSConfig.AccessKeyID == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.AccessKeyID must not be empty")
		}
		if config.HostAWSConfig.AccessKeySecret == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.AccessKeySecret must not be empty")
		}
		if config.HostAWSConfig.Region == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.Region must not be empty")
		}
		if config.HostAWSConfig.SessionToken == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.SessionToken must not be empty")
		}
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = config.K8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	versionedResources, err := newVersionedResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = config.G8sClient.ProviderV1alpha1().AWSConfigs("")

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

func newVersionedResources(config FrameworkConfig) (map[string][]framework.Resource, error) {
	var err error

	guestAWSConfig := awsclient.Config{
		AccessKeyID:     config.GuestAWSConfig.AccessKeyID,
		AccessKeySecret: config.GuestAWSConfig.AccessKeySecret,
		SessionToken:    config.GuestAWSConfig.SessionToken,
		Region:          config.GuestAWSConfig.Region,
	}

	hostAWSConfig := awsclient.Config{
		AccessKeyID:     config.HostAWSConfig.AccessKeyID,
		AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
		SessionToken:    config.HostAWSConfig.SessionToken,
		Region:          config.HostAWSConfig.Region,
	}

	awsClients := awsclient.NewClients(guestAWSConfig)
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

	awsHostClients := awsclient.NewClients(hostAWSConfig)

	var certWatcher *legacy.Service
	{
		certConfig := legacy.DefaultServiceConfig()
		certConfig.K8sClient = config.K8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = legacy.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keyWatcher *randomkeytpr.Service
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = config.K8sClient
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

	// ccServiceV3 is used by the s3objectv2.
	var ccServiceV3 *cloudconfigv3.CloudConfig
	{
		ccServiceConfig := cloudconfigv3.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV3, err = cloudconfigv3.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// ccServiceV4 is used by the s3objectv3.
	var ccServiceV4 *cloudconfigv4.CloudConfig
	{
		ccServiceConfig := cloudconfigv4.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccServiceV4, err = cloudconfigv4.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyResource framework.Resource
	{
		legacyConfig := legacyv2.DefaultConfig()
		legacyConfig.AwsConfig = guestAWSConfig
		legacyConfig.AwsHostConfig = hostAWSConfig
		legacyConfig.CertWatcher = certWatcher
		legacyConfig.CloudConfig = ccServiceV2
		legacyConfig.InstallationName = config.InstallationName
		legacyConfig.K8sClient = config.K8sClient
		legacyConfig.KeyWatcher = keyWatcher
		legacyConfig.Logger = config.Logger
		legacyConfig.PubKeyFile = config.PubKeyFile

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

		cloudformationConfig.InstallationName = config.InstallationName

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

	var s3BucketObjectResourceV2 framework.Resource
	{
		s3BucketObjectConfig := s3objectv2.DefaultConfig()
		s3BucketObjectConfig.AwsService = awsService
		s3BucketObjectConfig.Clients.S3 = awsClients.S3
		s3BucketObjectConfig.Clients.KMS = awsClients.KMS
		s3BucketObjectConfig.CloudConfig = ccServiceV3
		s3BucketObjectConfig.CertWatcher = certWatcher
		s3BucketObjectConfig.Logger = config.Logger
		s3BucketObjectConfig.RandomKeyWatcher = keyWatcher

		s3BucketObjectResourceV2, err = s3objectv2.New(s3BucketObjectConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketObjectResourceV3 framework.Resource
	{
		c := s3objectv3.DefaultConfig()
		c.AwsService = awsService
		c.Clients.S3 = awsClients.S3
		c.Clients.KMS = awsClients.KMS
		c.CloudConfig = ccServiceV4
		c.CertWatcher = certWatcher
		c.Logger = config.Logger
		c.RandomKeyWatcher = keyWatcher

		s3BucketObjectResourceV3, err = s3objectv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		namespaceConfig := namespacev2.DefaultConfig()

		namespaceConfig.K8sClient = config.K8sClient
		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespacev2.New(namespaceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		serviceConfig := servicev2.DefaultConfig()

		serviceConfig.K8sClient = config.K8sClient
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
		endpointsConfig.K8sClient = config.K8sClient
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

	// We create the list of resources and wrap each resource around some common
	// resources like metrics and retry resources.
	var resourcesV2_0_1 []framework.Resource
	{
		resourcesV2_0_1 = []framework.Resource{
			kmsKeyResource,
			s3BucketResource,
			s3BucketObjectResourceV3,
			cloudformationResource,
			namespaceResource,
			serviceResource,
			endpointsResource,
		}

		resourcesV2_0_1, err = metricsresource.Wrap(resourcesV2_0_1, metricsWrapConfig)
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
		"0.2.0": resourcesV2,
		"1.0.0": legacyResources,
		"2.0.0": resourcesV2,
		"2.0.1": resourcesV2_0_1,
		// 2.0.2 fixes missing region in host account credentials, the change only affects service/framework.go
		"2.0.2": resourcesV2_0_1,
	}

	return versionedResources, nil
}
