package v24

import (
	"context"
	"net"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"github.com/giantswarm/statusresource"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/cloudconfig"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v24/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/credential"
	"github.com/giantswarm/aws-operator/service/controller/v24/ebs"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter/kms"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter/vault"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/bridgezone"
	cloudformationresource "github.com/giantswarm/aws-operator/service/controller/v24/resource/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/cpi"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/ebsvolume"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/encryptionkey"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/loadbalancer"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/migration"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/namespace"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/service"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/stackoutput"
	"github.com/giantswarm/aws-operator/service/controller/v24/resource/workerasgname"
)

const (
	// minAllocatedSubnetMaskBits is the maximum size of guest subnet i.e.
	// smaller number here -> larger subnet per guest cluster. For now anything
	// under 16 doesn't make sense in here.
	minAllocatedSubnetMaskBits = 16
)

type ClusterResourceSetConfig struct {
	CertsSearcher      certs.Interface
	G8sClient          versioned.Interface
	HostAWSClients     aws.Clients
	HostAWSConfig      aws.Config
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               adapter.APIWhitelist
	EncrypterBackend           string
	GuestAvailabilityZones     []string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	GuestUpdateEnabled         bool
	IncludeTags                bool
	IgnitionPath               string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	DeleteLoggingBucket        bool
	OIDC                       cloudconfig.OIDCConfig
	ProjectName                string
	PublicRouteTables          string
	Route53Enabled             bool
	PodInfraContainerImage     string
	RegistryDomain             string
	SSOPublicKey               string
	VaultAddress               string
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

	if config.GuestSubnetMaskBits < minAllocatedSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestSubnetMaskBits (%d) must not be smaller than %d", config, config.GuestSubnetMaskBits, minAllocatedSubnetMaskBits)
	}
	if config.GuestPrivateSubnetMaskBits <= config.GuestSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestPrivateSubnetMaskBits (%d) must not be smaller or equal than %T.GuestSubnetMaskBits (%d)", config, config.GuestPrivateSubnetMaskBits, config, config.GuestSubnetMaskBits)
	}
	if config.GuestPublicSubnetMaskBits <= config.GuestSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestPublicSubnetMaskBits (%d) must not be smaller or equal than %T.GuestSubnetMaskBits (%d)", config, config.GuestPublicSubnetMaskBits, config, config.GuestSubnetMaskBits)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RandomKeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RandomkeysSearcher must not be empty", config)
	}

	if config.IgnitionPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", config)
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
	if config.SSOPublicKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.SSOPublicKey must not be empty", config)
	}

	var encrypterObject encrypter.Interface
	var encrypterRoleManager encrypter.RoleManager
	switch config.EncrypterBackend {
	case encrypter.VaultBackend:
		c := &vault.EncrypterConfig{
			Logger: config.Logger,

			Address: config.VaultAddress,
		}

		encrypterObject, err = vault.NewEncrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		encrypterRoleManager = encrypterObject.(encrypter.RoleManager)
	case encrypter.KMSBackend:
		c := &kms.EncrypterConfig{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		encrypterObject, err = kms.NewEncrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	default:
		return nil, microerror.Maskf(invalidConfigError, "unknown encrypter backend %q", config.EncrypterBackend)
	}

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
			Encrypter: encrypterObject,
			Logger:    config.Logger,

			IgnitionPath:           config.IgnitionPath,
			OIDC:                   config.OIDC,
			PodInfraContainerImage: config.PodInfraContainerImage,
			RegistryDomain:         config.RegistryDomain,
			SSOPublicKey:           config.SSOPublicKey,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var accountIDResource controller.Resource
	{
		c := accountid.Config{
			Logger: config.Logger,
		}

		accountIDResource, err = accountid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var asgStatusResource controller.Resource
	{
		c := asgstatus.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		asgStatusResource, err = asgstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionKeyResource controller.Resource
	{
		c := encryptionkey.Config{
			Encrypter: encrypterObject,
			Logger:    config.Logger,
		}

		encryptionKeyResource, err = encryptionkey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
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

	var ipamResource controller.Resource
	{
		c := ipam.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			AllocatedSubnetMaskBits: config.GuestSubnetMaskBits,
			AvailabilityZones:       config.GuestAvailabilityZones,
			NetworkRange:            config.IPAMNetworkRange,
		}

		ipamResource, err = ipam.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var bridgeZoneResource controller.Resource
	{
		c := bridgezone.Config{
			HostAWSConfig: config.HostAWSConfig,
			HostRoute53:   config.HostAWSClients.Route53,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		bridgeZoneResource, err = bridgezone.New(c)
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

	var s3ObjectResource controller.Resource
	{
		c := s3object.Config{
			CertsSearcher:      config.CertsSearcher,
			CloudConfig:        cloudConfig,
			Logger:             config.Logger,
			RandomKeysSearcher: config.RandomKeysSearcher,
		}

		ops, err := s3object.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		s3ObjectResource, err = toCRUDResource(config.Logger, ops)
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
			APIWhitelist: adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			EncrypterRoleManager: encrypterRoleManager,
			G8sClient:            config.G8sClient,
			HostClients: &adapter.Clients{
				EC2:            config.HostAWSClients.EC2,
				IAM:            config.HostAWSClients.IAM,
				STS:            config.HostAWSClients.STS,
				CloudFormation: config.HostAWSClients.CloudFormation,
			},
			Logger: config.Logger,

			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			EncrypterBackend:           config.EncrypterBackend,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			InstallationName:           config.InstallationName,
			PublicRouteTables:          config.PublicRouteTables,
			Route53Enabled:             config.Route53Enabled,
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

	var cpiResource controller.Resource
	{
		c := cpi.Config{
			HostClients: &adapter.Clients{
				EC2:            config.HostAWSClients.EC2,
				IAM:            config.HostAWSClients.IAM,
				STS:            config.HostAWSClients.STS,
				CloudFormation: config.HostAWSClients.CloudFormation,
			},
			Logger: config.Logger,

			InstallationName: config.InstallationName,
			Route53Enabled:   config.Route53Enabled,
		}

		cpiResource, err = cpi.New(c)
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

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,

			CertID: certs.APICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResource controller.Resource
	{
		c := statusresource.ResourceConfig{
			ClusterEndpointFunc:      key.ToClusterEndpoint,
			ClusterIDFunc:            key.ToClusterID,
			ClusterStatusFunc:        key.ToClusterStatus,
			NodeCountFunc:            key.ToNodeCount,
			Logger:                   config.Logger,
			RESTClient:               config.G8sClient.ProviderV1alpha1().RESTClient(),
			TenantCluster:            tenantCluster,
			VersionBundleVersionFunc: key.ToVersionBundleVersion,
		}

		statusResource, err = statusresource.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var stackOutputResource controller.Resource
	{
		c := stackoutput.Config{
			EC2:    config.HostAWSClients.EC2,
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		stackOutputResource, err = stackoutput.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerASGNameResource controller.Resource
	{
		c := workerasgname.ResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		workerASGNameResource, err = workerasgname.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		accountIDResource,
		stackOutputResource,
		workerASGNameResource,
		asgStatusResource,
		statusResource,
		migrationResource,
		ipamResource,
		bridgeZoneResource,
		encryptionKeyResource,
		s3BucketResource,
		s3ObjectResource,
		loadBalancerResource,
		ebsVolumeResource,
		cpiResource,
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
		c := metricsresource.WrapConfig{}

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
