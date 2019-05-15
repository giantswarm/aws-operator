package v27

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"github.com/giantswarm/statusresource"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/adapter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/credential"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/detection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter/kms"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter/vault"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/bridgezone"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/cpf"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/cpi"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/ebsvolume"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/encryption"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/loadbalancer"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/namespace"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/natgatewayaddresses"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/peerrolearn"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/routetable"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/service"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/tccp"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/tccpsubnet"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/vpccidr"
	"github.com/giantswarm/aws-operator/service/network"
)

const (
	// minAllocatedSubnetMaskBits is the maximum size of guest subnet i.e.
	// smaller number here -> larger subnet per guest cluster. For now anything
	// under 16 doesn't make sense in here.
	minAllocatedSubnetMaskBits = 16
)

type ClusterResourceSetConfig struct {
	CertsSearcher          certs.Interface
	ControlPlaneAWSClients aws.Clients
	CMAClient              clientset.Interface
	G8sClient              versioned.Interface
	HostAWSConfig          aws.Config
	K8sClient              kubernetes.Interface
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
	VPCPeerID                  string
}

func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
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

	var encryptionResource controller.Resource
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

	var ipamResource controller.Resource
	{
		c := ipam.Config{
			CMAClient:        config.CMAClient,
			G8sClient:        config.G8sClient,
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

	var bridgeZoneResource controller.Resource
	{
		c := bridgezone.Config{
			HostAWSConfig: config.HostAWSConfig,
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

	var tccpResource controller.Resource
	{
		c := tccp.Config{
			APIWhitelist:         config.APIWhitelist,
			EncrypterRoleManager: encrypterRoleManager,
			Logger:               config.Logger,

			Detection:          detectionService,
			EncrypterBackend:   config.EncrypterBackend,
			InstallationName:   config.InstallationName,
			InstanceMonitoring: config.AdvancedMonitoringEC2,
			PublicRouteTables:  config.RouteTables,
			Route53Enabled:     config.Route53Enabled,
			VPCPeerID:          config.VPCPeerID,
		}

		tccpResource, err = tccp.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpOutputsResource controller.Resource
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

	var tccpSubnetResource controller.Resource
	{
		c := tccpsubnet.Config{
			Logger: config.Logger,
		}

		tccpSubnetResource, err = tccpsubnet.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpfResource controller.Resource
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

	var cpiResource controller.Resource
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

	var natGatewayAddressesResource controller.Resource
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

	var peerRoleARNResource controller.Resource
	{
		c := peerrolearn.Config{
			Logger: config.Logger,
		}

		peerRoleARNResource, err = peerrolearn.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var routeTableResource controller.Resource
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
			ClusterEndpointFunc:      legacykey.ToClusterEndpoint,
			ClusterIDFunc:            legacykey.ToClusterID,
			ClusterStatusFunc:        legacykey.ToClusterStatus,
			NodeCountFunc:            legacykey.ToNodeCount,
			Logger:                   config.Logger,
			RESTClient:               config.G8sClient.ProviderV1alpha1().RESTClient(),
			TenantCluster:            tenantCluster,
			VersionBundleVersionFunc: legacykey.ToVersionBundleVersion,
		}

		statusResource, err = statusresource.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vpcCIDRResource controller.Resource
	{
		c := vpccidr.Config{
			Logger: config.Logger,

			VPCPeerID: config.VPCPeerID,
		}

		vpcCIDRResource, err = vpccidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		accountIDResource,
		natGatewayAddressesResource,
		peerRoleARNResource,
		routeTableResource,
		vpcCIDRResource,
		tccpOutputsResource,
		tccpSubnetResource,
		asgStatusResource,
		statusResource,
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
		customObject, err := legacykey.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if legacykey.ClusterVersion(customObject) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		var tenantClusterAWSClients aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient, obj)
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
