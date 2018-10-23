package v12patch1

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/legacycerts/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/cloudconfig"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v12patch1/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/credential"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/ebs"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/key"
	cloudformationresource "github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/ebsvolume"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/kmskey"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/loadbalancer"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/migration"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/namespace"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/resource/service"
)

type ClusterResourceSetConfig struct {
	CertsSearcher      legacy.Searcher
	G8sClient          versioned.Interface
	HostAWSConfig      aws.Config
	HostAWSClients     aws.Clients
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	AccessLogsExpiration   int
	AdvancedMonitoringEC2  bool
	APIWhitelist           adapter.APIWhitelist
	GuestUpdateEnabled     bool
	IncludeTags            bool
	InstallationName       string
	DeleteLoggingBucket    bool
	OIDC                   cloudconfig.OIDCConfig
	ProjectName            string
	Route53Enabled         bool
	PodInfraContainerImage string
}

func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HostAWSConfig.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSConfig.AccessKeyID must not be empty", config)
	}
	if config.HostAWSConfig.AccessKeySecret == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSConfig.AccessKeySecret must not be empty", config)
	}
	if config.HostAWSConfig.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSConfig.Region must not be empty", config)
	}
	if config.HostAWSClients.CloudFormation == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSClients.CloudFormation must not be empty", config)
	}
	if config.HostAWSClients.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSClients.EC2 must not be empty", config)
	}
	if config.HostAWSClients.IAM == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSClients.IAM must not be empty", config)
	}
	if config.HostAWSClients.Route53 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSClients.Route53 must not be empty", config)
	}

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RandomkeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RandomkeysSearcher must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}
	if config.APIWhitelist.Enabled && config.APIWhitelist.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.SubnetList must not be empty when %T.APIWhitelist is enabled", config)
	}

	var migrationResource controller.Resource
	{
		c := migration.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		migrationResource, err = migration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kmsKeyResource controller.Resource
	{
		c := kmskey.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		ops, err := kmskey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		kmsKeyResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketResource controller.Resource
	{
		c := s3bucket.Config{
			Logger: config.Logger,

			AccessLogsExpiration: config.AccessLogsExpiration,
			DeleteLoggingBucket:  config.DeleteLoggingBucket,
			IncludeTags:          config.IncludeTags,
			InstallationName:     config.InstallationName,
		}

		ops, err := s3bucket.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		s3BucketResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketObjectResource controller.Resource
	{
		c := s3object.Config{
			CertWatcher:       config.CertsSearcher,
			Logger:            config.Logger,
			RandomKeySearcher: config.RandomkeysSearcher,
		}

		ops, err := s3object.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		s3BucketObjectResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var loadBalancerResource controller.Resource
	{
		c := loadbalancer.Config{
			Logger: config.Logger,
		}

		loadBalancerResource, err = loadbalancer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ebsVolumeResource controller.Resource
	{
		c := ebsvolume.Config{
			Logger: config.Logger,
		}

		ebsVolumeResource, err = ebsvolume.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudformationResource controller.Resource
	{
		c := cloudformationresource.Config{
			HostClients: &adapter.Clients{
				EC2:            config.HostAWSClients.EC2,
				IAM:            config.HostAWSClients.IAM,
				STS:            config.HostAWSClients.STS,
				CloudFormation: config.HostAWSClients.CloudFormation,
			},
			Logger: config.Logger,
			APIWhitelist: adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},

			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			InstallationName:      config.InstallationName,
			Route53Enabled:        config.Route53Enabled,
		}

		ops, err := cloudformationresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		cloudformationResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource controller.Resource
	{
		c := namespace.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		namespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource controller.Resource
	{
		c := service.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := service.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		serviceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var endpointsResource controller.Resource
	{
		c := endpoints.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := endpoints.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		endpointsResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		migrationResource,
		kmsKeyResource,
		s3BucketResource,
		s3BucketObjectResource,
		loadBalancerResource,
		ebsVolumeResource,
		cloudformationResource,
		namespaceResource,
		serviceResource,
		endpointsResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(customObject) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		if config.GuestUpdateEnabled {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}

		var awsClient aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient, obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			c := config.HostAWSConfig
			c.RoleARN = arn

			awsClient, err = aws.NewClients(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		var awsService *awsservice.Service
		{
			c := awsservice.Config{
				Clients: awsservice.Clients{
					KMS: awsClient.KMS,
					STS: awsClient.STS,
				},
				Logger: config.Logger,
			}

			awsService, err = awsservice.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		var cloudConfig *cloudconfig.CloudConfig
		{
			c := cloudconfig.Config{
				KMSClient: awsClient.KMS,
				Logger:    config.Logger,

				OIDC: config.OIDC,
				PodInfraContainerImage: config.PodInfraContainerImage,
			}

			cloudConfig, err = cloudconfig.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		var ebsService ebs.Interface
		{
			c := ebs.Config{
				Client: awsClient.EC2,
				Logger: config.Logger,
			}
			ebsService, err = ebs.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		var cloudFormationService *cloudformationservice.CloudFormation
		{
			c := cloudformationservice.Config{
				Client: awsClient.CloudFormation,
			}

			cloudFormationService, err = cloudformationservice.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		c := controllercontext.Context{
			AWSClient:      awsClient,
			AWSService:     awsService,
			CloudConfig:    cloudConfig,
			CloudFormation: *cloudFormationService,
			EBSService:     ebsService,
		}
		ctx = controllercontext.NewContext(ctx, c)

		return ctx, nil
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
