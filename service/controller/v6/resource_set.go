package v6

import (
	"context"

	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v6/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v6/key"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/cloudformation/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/ebsvolume"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/kmskey"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/loadbalancer"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/namespace"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/v6/resource/service"
)

type ResourceSetConfig struct {
	CertsSearcher      legacy.Searcher
	GuestAWSClients    aws.Clients
	HostAWSClients     aws.Clients
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher *randomkeytpr.Service

	GuestUpdateEnabled bool
	InstallationName   string
	OIDC               cloudconfig.OIDCConfig
	ProjectName        string
}

func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
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

	var awsService *awsservice.Service
	{
		c := awsservice.Config{
			Clients: awsservice.Clients{
				KMS: config.GuestAWSClients.KMS,
				STS: config.GuestAWSClients.STS,
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
			Logger: config.Logger,
			OIDC:   config.OIDC,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kmsKeyResource controller.CRUDResourceOps
	{
		c := kmskey.Config{
			Clients: kmskey.Clients{
				KMS: config.GuestAWSClients.KMS,
			},
			Logger: config.Logger,
		}

		kmsKeyResource, err = kmskey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketResource controller.CRUDResourceOps
	{
		c := s3bucket.Config{
			AwsService: awsService,
			Clients: s3bucket.Clients{
				S3: config.GuestAWSClients.S3,
			},
			Logger: config.Logger,
		}

		s3BucketResource, err = s3bucket.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketObjectResource controller.CRUDResourceOps
	{
		c := s3object.Config{
			AwsService: awsService,
			Clients: s3object.Clients{
				S3:  config.GuestAWSClients.S3,
				KMS: config.GuestAWSClients.KMS,
			},
			CloudConfig:      cloudConfig,
			CertWatcher:      config.CertsSearcher,
			Logger:           config.Logger,
			RandomKeyWatcher: config.RandomkeysSearcher,
		}

		s3BucketObjectResource, err = s3object.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var loadBalancerResource controller.CRUDResourceOps
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

	var ebsVolumeResource controller.CRUDResourceOps
	{
		c := ebsvolume.Config{
			Clients: ebsvolume.Clients{
				EC2: config.GuestAWSClients.EC2,
			},
			Logger: config.Logger,
		}

		ebsVolumeResource, err = ebsvolume.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudformationResource controller.CRUDResourceOps
	{
		c := cloudformation.Config{
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
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		cloudformationResource, err = cloudformation.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource controller.CRUDResourceOps
	{
		c := namespace.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		namespaceResource, err = namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource controller.CRUDResourceOps
	{
		c := service.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		serviceResource, err = service.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var endpointsResource controller.CRUDResourceOps
	{
		c := endpoints.Config{
			Clients: endpoints.Clients{
				EC2: config.GuestAWSClients.EC2,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		endpointsResource, err = endpoints.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []controller.Resource
	ops := []controller.CRUDResourceOps{
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
	for _, o := range ops {
		c := controller.CRUDResourceConfig{
			Logger: config.Logger,
			Ops:    o,
		}

		r, err := controller.NewCRUDResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		resources = append(resources, r)
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
