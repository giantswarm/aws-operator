package controller

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/crud"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys"
	"github.com/giantswarm/statusresource"
	"github.com/giantswarm/tenantcluster"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/adapter"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/internal/detection"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter/vault"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/resource/bridgezone"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanupsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpf"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpi"
	"github.com/giantswarm/aws-operator/service/controller/resource/ebsvolume"
	"github.com/giantswarm/aws-operator/service/controller/resource/encryption"
	"github.com/giantswarm/aws-operator/service/controller/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/resource/loadbalancer"
	"github.com/giantswarm/aws-operator/service/controller/resource/migration"
	"github.com/giantswarm/aws-operator/service/controller/resource/namespace"
	"github.com/giantswarm/aws-operator/service/controller/resource/natgatewayaddresses"
	"github.com/giantswarm/aws-operator/service/controller/resource/peerrolearn"
	"github.com/giantswarm/aws-operator/service/controller/resource/routetable"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/resource/secretfinalizer"
	"github.com/giantswarm/aws-operator/service/controller/resource/service"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsubnet"
	"github.com/giantswarm/aws-operator/service/controller/resource/vpc"
	"github.com/giantswarm/aws-operator/service/internal/credential"
	"github.com/giantswarm/aws-operator/service/internal/network"
)

const (
	// minAllocatedSubnetMaskBits is the maximum size of guest subnet i.e.
	// smaller number here -> larger subnet per guest cluster. For now anything
	// under 16 doesn't make sense in here.
	minAllocatedSubnetMaskBits = 16
)

type clusterResourceSetConfig struct {
	CertsSearcher          certs.Interface
	ControlPlaneAWSClients aws.Clients
	HostAWSConfig          aws.Config
	K8sClient              k8sclient.Interface
	Logger                 micrologger.Logger
	NetworkAllocator       network.Allocator
	RandomKeysSearcher     randomkeys.Interface

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               adapter.APIWhitelist
	EncrypterBackend           string
	GuestAvailabilityZones     []string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	IncludeTags                bool
	IgnitionPath               string
	ImagePullProgressDeadline  string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	DeleteLoggingBucket        bool
	OIDC                       cloudconfig.OIDCConfig
	ProjectName                string
	Route53Enabled             bool
	RouteTables                string
	PodInfraContainerImage     string
	RegistryDomain             string
	SSOPublicKey               string
	VaultAddress               string
}

func newClusterResourceSet(config clusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.GuestSubnetMaskBits < minAllocatedSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestSubnetMaskBits (%d) must not be smaller than %d", config, config.GuestSubnetMaskBits, minAllocatedSubnetMaskBits)
	}
	if config.GuestPrivateSubnetMaskBits <= config.GuestSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestPrivateSubnetMaskBits (%d) must not be smaller or equal than %T.GuestSubnetMaskBits (%d)", config, config.GuestPrivateSubnetMaskBits, config, config.GuestSubnetMaskBits)
	}
	if config.GuestPublicSubnetMaskBits <= config.GuestSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestPublicSubnetMaskBits (%d) must not be smaller or equal than %T.GuestSubnetMaskBits (%d)", config, config.GuestPublicSubnetMaskBits, config, config.GuestSubnetMaskBits)
	}
	if config.IgnitionPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", config)
	}
	if config.ImagePullProgressDeadline == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ImagePullProgressDeadline must not be empty", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}
	if config.APIWhitelist.Public.Enabled && config.APIWhitelist.Public.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Public.SubnetList must not be empty when %T.APIWhitelist.Public is enabled", config, config)
	}
	if config.APIWhitelist.Private.Enabled && config.APIWhitelist.Private.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Private.SubnetList must not be empty when %T.APIWhitelist.Private is enabled", config, config)
	}
	if config.SSOPublicKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.SSOPublicKey must not be empty", config)
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

			IgnitionPath:              config.IgnitionPath,
			ImagePullProgressDeadline: config.ImagePullProgressDeadline,
			OIDC:                      config.OIDC,
			PodInfraContainerImage:    config.PodInfraContainerImage,
			RegistryDomain:            config.RegistryDomain,
			SSOPublicKey:              config.SSOPublicKey,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var detectionService *detection.Detection
	{
		c := detection.Config{
			Logger: config.Logger,
		}

		detectionService, err = detection.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var accountIDResource resource.Interface
	{
		c := accountid.Config{
			Logger: config.Logger,
		}

		accountIDResource, err = accountid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var asgStatusResource resource.Interface
	{
		c := asgstatus.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		asgStatusResource, err = asgstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cleanupSecurityGroups resource.Interface
	{
		c := cleanupsecuritygroups.Config{
			Logger: config.Logger,
		}

		cleanupSecurityGroups, err = cleanupsecuritygroups.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionResource resource.Interface
	{
		c := encryption.Config{
			Encrypter: encrypterObject,
			Logger:    config.Logger,
		}

		encryptionResource, err = encryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var migrationResource resource.Interface
	{
		c := migration.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		migrationResource, err = migration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ipamResource resource.Interface
	{
		c := ipam.Config{
			G8sClient:        config.K8sClient.G8sClient(),
			Logger:           config.Logger,
			NetworkAllocator: config.NetworkAllocator,

			AllocatedSubnetMaskBits: config.GuestSubnetMaskBits,
			AvailabilityZones:       config.GuestAvailabilityZones,
			NetworkRange:            config.IPAMNetworkRange,
		}

		ipamResource, err = ipam.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var bridgeZoneResource resource.Interface
	{
		c := bridgezone.Config{
			HostAWSConfig: config.HostAWSConfig,
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		bridgeZoneResource, err = bridgezone.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3BucketResource resource.Interface
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

	var s3ObjectResource resource.Interface
	{
		c := s3object.Config{
			CertsSearcher:      config.CertsSearcher,
			CloudConfig:        cloudConfig,
			G8sClient:          config.K8sClient.G8sClient(),
			Logger:             config.Logger,
			RandomKeysSearcher: config.RandomKeysSearcher,
			RegistryDomain:     config.RegistryDomain,
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

	var loadBalancerResource resource.Interface
	{
		c := loadbalancer.Config{
			Logger: config.Logger,
		}

		loadBalancerResource, err = loadbalancer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ebsVolumeResource resource.Interface
	{
		c := ebsvolume.Config{
			Logger: config.Logger,
		}

		ebsVolumeResource, err = ebsvolume.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpResource resource.Interface
	{
		c := tccp.Config{
			APIWhitelist: adapter.APIWhitelist{
				Private: config.APIWhitelist.Private,
				Public:  config.APIWhitelist.Public,
			},
			EncrypterRoleManager: encrypterRoleManager,
			G8sClient:            config.K8sClient.G8sClient(),
			Logger:               config.Logger,

			Detection:          detectionService,
			EncrypterBackend:   config.EncrypterBackend,
			InstallationName:   config.InstallationName,
			InstanceMonitoring: config.AdvancedMonitoringEC2,
			PublicRouteTables:  config.RouteTables,
			Route53Enabled:     config.Route53Enabled,
		}

		tccpResource, err = tccp.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpOutputsResource resource.Interface
	{
		c := tccpoutputs.Config{
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		tccpOutputsResource, err = tccpoutputs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSubnetResource resource.Interface
	{
		c := tccpsubnet.Config{
			Logger: config.Logger,
		}

		tccpSubnetResource, err = tccpsubnet.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpfResource resource.Interface
	{
		c := cpf.Config{
			Logger: config.Logger,

			EncrypterBackend: config.EncrypterBackend,
			InstallationName: config.InstallationName,
			Route53Enabled:   config.Route53Enabled,
		}

		cpfResource, err = cpf.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpiResource resource.Interface
	{
		c := cpi.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		cpiResource, err = cpi.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource resource.Interface
	{
		c := namespace.Config{
			K8sClient: config.K8sClient.K8sClient(),
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

	var natGatewayAddressesResource resource.Interface
	{
		c := natgatewayaddresses.Config{
			Logger: config.Logger,

			Installation: config.InstallationName,
		}

		natGatewayAddressesResource, err = natgatewayaddresses.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var peerRoleARNResource resource.Interface
	{
		c := peerrolearn.Config{
			Logger: config.Logger,
		}

		peerRoleARNResource, err = peerrolearn.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var routeTableResource resource.Interface
	{
		c := routetable.Config{
			Logger: config.Logger,

			Names: strings.Split(config.RouteTables, ","),
		}

		routeTableResource, err = routetable.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var secretFinalizerResource resource.Interface
	{
		c := secretfinalizer.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		secretFinalizerResource, err = secretfinalizer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource resource.Interface
	{
		c := service.Config{
			K8sClient: config.K8sClient.K8sClient(),
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

	var endpointsResource resource.Interface
	{
		c := endpoints.Config{
			K8sClient: config.K8sClient.K8sClient(),
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
			K8sClient: config.K8sClient.K8sClient(),
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

	var statusResource resource.Interface
	{
		c := statusresource.ResourceConfig{
			ClusterEndpointFunc:      key.ToClusterEndpoint,
			ClusterIDFunc:            key.ToClusterID,
			ClusterStatusFunc:        key.ToClusterStatus,
			NodeCountFunc:            key.ToNodeCount,
			Logger:                   config.Logger,
			RESTClient:               config.K8sClient.G8sClient().ProviderV1alpha1().RESTClient(),
			TenantCluster:            tenantCluster,
			VersionBundleVersionFunc: key.ToVersionBundleVersion,
		}

		statusResource, err = statusresource.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vpcCIDRResource resource.Interface
	{
		c := vpc.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		vpcCIDRResource, err = vpc.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		accountIDResource,
		natGatewayAddressesResource,
		peerRoleARNResource,
		routeTableResource,
		vpcCIDRResource,
		tccpOutputsResource,
		tccpSubnetResource,
		asgStatusResource,
		statusResource,
		migrationResource,
		ipamResource,
		bridgeZoneResource,
		encryptionResource,
		s3BucketResource,
		s3ObjectResource,
		loadBalancerResource,
		ebsVolumeResource,
		cpiResource,
		tccpResource,
		cpfResource,
		namespaceResource,
		serviceResource,
		endpointsResource,
		secretFinalizerResource,
		cleanupSecurityGroups,
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

		if key.VersionBundleVersion(customObject) == project.BundleVersion() {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		var tenantClusterAWSClients aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient.K8sClient(), obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			c := config.HostAWSConfig
			c.RoleARN = arn

			tenantClusterAWSClients, err = aws.NewClients(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		c := controllercontext.Context{
			Client: controllercontext.ContextClient{
				ControlPlane: controllercontext.ContextClientControlPlane{
					AWS: config.ControlPlaneAWSClients,
				},
				TenantCluster: controllercontext.ContextClientTenantCluster{
					AWS: tenantClusterAWSClients,
				},
			},
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

func toCRUDResource(logger micrologger.Logger, ops crud.Interface) (resource.Interface, error) {
	c := crud.ResourceConfig{
		CRUD:   ops,
		Logger: logger,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
