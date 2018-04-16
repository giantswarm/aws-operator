package v10

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/adapter"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/cloudconfig"
	cloudformationservice "github.com/giantswarm/aws-operator/service/awsconfig/v10/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/ebs"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
	cloudformationresource "github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/ebsvolume"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/kmskey"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/loadbalancer"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/namespace"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/s3object"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/resource/service"
)

const (
	ResourceRetries uint64 = 3
)

type ClusterResourceSetConfig struct {
	CertsSearcher      legacy.Searcher
	GuestAWSClients    aws.Clients
	HostAWSClients     aws.Clients
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	GuestUpdateEnabled   bool
	InstallationName     string
	AccessLogsExpiration int
	OIDC                 cloudconfig.OIDCConfig
	ProjectName          string
}

func NewClusterResourceSet(config ClusterResourceSetConfig) (*framework.ResourceSet, error) {
	var err error

	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertsSearcher must not be empty")
	}
	if config.GuestAWSClients.CloudFormation == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSClients.CloudFormation must not be empty")
	}
	if config.GuestAWSClients.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSClients.EC2 must not be empty")
	}
	if config.GuestAWSClients.ELB == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSClients.ELB must not be empty")
	}
	if config.GuestAWSClients.IAM == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSClients.IAM must not be empty")
	}
	if config.GuestAWSClients.KMS == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSClients.KMS must not be empty")
	}
	if config.GuestAWSClients.S3 == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSClients.S3 must not be empty")
	}
	if config.HostAWSClients.CloudFormation == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.HostAWSClients.CloudFormation must not be empty")
	}
	if config.HostAWSClients.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.HostAWSClients.EC2 must not be empty")
	}
	if config.HostAWSClients.IAM == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.HostAWSClients.IAM must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RandomkeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RandomkeysSearcher must not be empty")
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.InstallationName must not be empty")
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
	}
	if config.AccessLogsExpiration == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.AccessLogsExpiration must not be empty")
	}

	var awsService *awsservice.Service
	{
		c := awsservice.Config{
			Clients: awsservice.Clients{
				IAM: config.GuestAWSClients.IAM,
				KMS: config.GuestAWSClients.KMS,
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
			KMSClient: config.GuestAWSClients.KMS,
			Logger:    config.Logger,

			OIDC: config.OIDC,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudFormationService *cloudformationservice.CloudFormation
	{
		c := cloudformationservice.Config{
			Client: config.GuestAWSClients.CloudFormation,
		}

		cloudFormationService, err = cloudformationservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ebsService ebs.Interface
	{
		c := ebs.Config{
			Client: config.GuestAWSClients.EC2,
			Logger: config.Logger,
		}
		ebsService, err = ebs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kmsKeyResource framework.Resource
	{
		c := kmskey.Config{
			Clients: kmskey.Clients{
				KMS: config.GuestAWSClients.KMS,
			},
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

	var s3BucketResource framework.Resource
	{
		c := s3bucket.Config{
			AwsService: awsService,
			Clients: s3bucket.Clients{
				S3: config.GuestAWSClients.S3,
			},
			Logger: config.Logger,

			InstallationName:     config.InstallationName,
			AccessLogsExpiration: config.AccessLogsExpiration,
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

	var s3BucketObjectResource framework.Resource
	{
		c := s3object.Config{
			AwsService: awsService,
			Clients: s3object.Clients{
				S3:  config.GuestAWSClients.S3,
				KMS: config.GuestAWSClients.KMS,
			},
			CloudConfig:       cloudConfig,
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

	var loadBalancerResource framework.Resource
	{
		c := loadbalancer.Config{
			Clients: loadbalancer.Clients{
				ELB: config.GuestAWSClients.ELB,
			},
			Logger: config.Logger,
		}

		loadBalancerResource, err = loadbalancer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ebsVolumeResource framework.Resource
	{
		c := ebsvolume.Config{
			Logger:  config.Logger,
			Service: ebsService,
		}

		ops, err := ebsvolume.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ebsVolumeResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudformationResource framework.Resource
	{
		c := cloudformationresource.Config{
			Clients: &adapter.Clients{
				EC2:            config.GuestAWSClients.EC2,
				CloudFormation: config.GuestAWSClients.CloudFormation,
				IAM:            config.GuestAWSClients.IAM,
				KMS:            config.GuestAWSClients.KMS,
				ELB:            config.GuestAWSClients.ELB,
			},
			HostClients: &adapter.Clients{
				EC2:            config.HostAWSClients.EC2,
				IAM:            config.HostAWSClients.IAM,
				CloudFormation: config.HostAWSClients.CloudFormation,
			},
			Logger:  config.Logger,
			Service: cloudFormationService,

			InstallationName: config.InstallationName,
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

	var namespaceResource framework.Resource
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

	var serviceResource framework.Resource
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

	var endpointsResource framework.Resource
	{
		c := endpoints.Config{
			Clients: endpoints.Clients{
				EC2: config.GuestAWSClients.EC2,
			},
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

	resources := []framework.Resource{
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
			BackOffFactory: func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) },
			Logger:         config.Logger,
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

		return ctx, nil
	}

	var resourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func toCRUDResource(logger micrologger.Logger, ops framework.CRUDResourceOps) (*framework.CRUDResource, error) {
	c := framework.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := framework.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
