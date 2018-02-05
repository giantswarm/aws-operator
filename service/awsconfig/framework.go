package awsconfig

import (
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
	"github.com/giantswarm/aws-operator/service/awsconfig/v2"
	cloudconfigv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/cloudconfig"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
	legacyv2 "github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/legacy"
	"github.com/giantswarm/aws-operator/service/awsconfig/v3"
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
	// config.GuestAWSConfig.SessionToken is optional.
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
		// config.HostAWSConfig.SessionToken is optional.
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

	resourceRouter, err := newResourceRouter(config)
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

	var crdFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.CRD = v1alpha1.NewAWSConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.Logger = config.Logger
		c.ResourceRouter = resourceRouter

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}

func newResourceRouter(config FrameworkConfig) (*framework.ResourceRouter, error) {
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

	handlesFuncFunc := func(handledVersionBundles []string) func(obj interface{}) bool {
		return func(obj interface{}) bool {
			awsConfig, err := key.ToCustomObject(obj)
			if err != nil {
				return false
			}
			versionBundleVersion := key.VersionBundleVersion(awsConfig)

			for _, v := range handledVersionBundles {
				if versionBundleVersion == v {
					return true
				}
			}

			return false
		}
	}

	var resourceSetV1 *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{
			Handles: handlesFuncFunc([]string{
				"",
				"0.1.0",
				"1.0.0",
			}),
			Logger:    config.Logger,
			Resources: legacyResources,
		}

		resourceSetV1, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV2 *framework.ResourceSet
	{
		c := v2.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"0.2.0",
				"2.0.0",
			},
			InstallationName: config.InstallationName,
			ProjectName:      config.Name,
		}

		resourceSetV2, err = v2.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV3 *framework.ResourceSet
	{
		c := v3.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"2.0.1",
				// 2.0.2 fixes missing region in host account credentials, the change only affects service/framework.go
				"2.0.2",
			},
			InstallationName: config.InstallationName,
			ProjectName:      config.Name,
		}

		resourceSetV3, err = v3.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{
			ResourceSets: []*framework.ResourceSet{
				resourceSetV1,
				resourceSetV2,
				resourceSetV3,
			},
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceRouter, nil
}
