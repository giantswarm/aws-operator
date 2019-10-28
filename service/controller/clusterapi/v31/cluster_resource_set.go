package v31

import (
	"context"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/adapter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/bridgezone"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanupebsvolumes"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanuploadbalancers"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanupsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/cproutetables"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpvpccidr"
	"github.com/giantswarm/aws-operator/service/controller/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/resource/natgatewayaddresses"
	"github.com/giantswarm/aws-operator/service/controller/resource/peerrolearn"
	"github.com/giantswarm/aws-operator/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/resource/secretfinalizer"
	"github.com/giantswarm/aws-operator/service/controller/resource/service"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpencryption"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpf"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpi"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcid"
)

func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var encrypterObject encrypter.Interface
	{
		encrypterObject, err = newEncrypterObject(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encrypterRoleManager encrypter.RoleManager
	{
		encrypterRoleManager, err = newEncrypterRoleManager(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpCloudConfig *cloudconfig.TCCP
	{
		c := cloudconfig.TCCPConfig{
			Config: cloudconfig.Config{
				Encrypter: encrypterObject,
				Logger:    config.Logger,

				CalicoCIDR:                config.CalicoCIDR,
				CalicoMTU:                 config.CalicoMTU,
				CalicoSubnet:              config.CalicoSubnet,
				ClusterIPRange:            config.ClusterIPRange,
				DockerDaemonCIDR:          config.DockerDaemonCIDR,
				IgnitionPath:              config.IgnitionPath,
				ImagePullProgressDeadline: config.ImagePullProgressDeadline,
				NetworkSetupDockerImage:   config.NetworkSetupDockerImage,
				PodInfraContainerImage:    config.PodInfraContainerImage,
				RegistryDomain:            config.RegistryDomain,
				SSHUserList:               config.SSHUserList,
				SSOPublicKey:              config.SSOPublicKey,
			},
		}

		tccpCloudConfig, err = cloudconfig.NewTCCP(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpChangeDetection *changedetection.TCCP
	{
		c := changedetection.TCCPConfig{
			Logger: config.Logger,
		}

		tccpChangeDetection, err = changedetection.NewTCCP(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterChecker *ipam.ClusterChecker
	{
		c := ipam.ClusterCheckerConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		clusterChecker, err = ipam.NewClusterChecker(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var subnetCollector *ipam.SubnetCollector
	{
		c := ipam.SubnetCollectorConfig{
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			NetworkRange: config.IPAMNetworkRange,
		}

		subnetCollector, err = ipam.NewSubnetCollector(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterPersister *ipam.ClusterPersister
	{
		c := ipam.ClusterPersisterConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		clusterPersister, err = ipam.NewClusterPersister(c)
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

	var awsClientResource resource.Interface
	{
		c := awsclient.Config{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpAZsResource resource.Interface
	{
		c := tccpazs.Config{
			CMAClient:     config.CMAClient,
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,
		}

		tccpAZsResource, err = tccpazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpEncryptionResource resource.Interface
	{
		c := tccpencryption.Config{
			Encrypter: encrypterObject,
			Logger:    config.Logger,
		}

		tccpEncryptionResource, err = tccpencryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ipamResource resource.Interface
	{
		c := ipam.Config{
			Checker:   clusterChecker,
			Collector: subnetCollector,
			Locker:    config.Locker,
			Logger:    config.Logger,
			Persister: clusterPersister,

			AllocatedSubnetMaskBits: config.GuestSubnetMaskBits,
			NetworkRange:            config.IPAMNetworkRange,
			PrivateSubnetMaskBits:   config.GuestPrivateSubnetMaskBits,
			PublicSubnetMaskBits:    config.GuestPublicSubnetMaskBits,
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
			K8sClient:     config.K8sClient,
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
			CloudConfig:        tccpCloudConfig,
			LabelsFunc:         key.KubeletLabelsTCCP,
			Logger:             config.Logger,
			CMAClient:          config.CMAClient,
			PathFunc:           key.S3ObjectPathTCCP,
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

	var cleanupEBSVolumesResource resource.Interface
	{
		c := cleanupebsvolumes.Config{
			Logger: config.Logger,
		}

		cleanupEBSVolumesResource, err = cleanupebsvolumes.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cleanupLoadBalancersResource resource.Interface
	{
		c := cleanuploadbalancers.Config{
			Logger: config.Logger,
		}

		cleanupLoadBalancersResource, err = cleanuploadbalancers.New(c)
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

	var regionResource resource.Interface
	{
		c := region.Config{
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,
		}

		regionResource, err = region.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpResource resource.Interface
	{
		c := tccp.Config{
			CMAClient:            config.CMAClient,
			EncrypterRoleManager: encrypterRoleManager,
			Logger:               config.Logger,

			APIWhitelist: adapter.APIWhitelist{
				Private: config.APIWhitelist.Private,
				Public:  config.APIWhitelist.Public,
			},
			Detection:          tccpChangeDetection,
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

	var tccpSubnetsResource resource.Interface
	{
		c := tccpsubnets.Config{
			Logger: config.Logger,
		}

		tccpSubnetsResource, err = tccpsubnets.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpfResource resource.Interface
	{
		c := tccpf.Config{
			Logger: config.Logger,

			EncrypterBackend: config.EncrypterBackend,
			InstallationName: config.InstallationName,
			Route53Enabled:   config.Route53Enabled,
		}

		tccpfResource, err = tccpf.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpiResource resource.Interface
	{
		c := tccpi.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		tccpiResource, err = tccpi.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpVPCIDResource resource.Interface
	{
		c := tccpvpcid.Config{
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,
		}

		tccpVPCIDResource, err = tccpvpcid.New(c)
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

	var cpRouteTablesResource resource.Interface
	{
		c := cproutetables.Config{
			Logger: config.Logger,

			Names: strings.Split(config.RouteTables, ","),
		}

		cpRouteTablesResource, err = cproutetables.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var secretFinalizerResource resource.Interface
	{
		c := secretfinalizer.Config{
			K8sClient: config.K8sClient,
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

	var endpointsResource resource.Interface
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

	var vpcCIDRResource resource.Interface
	{
		c := cpvpccidr.Config{
			Logger: config.Logger,

			VPCPeerID: config.VPCPeerID,
		}

		vpcCIDRResource, err = cpvpccidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// All these resources only fetch information from remote APIs and put them
		// into the controller context.
		awsClientResource,
		accountIDResource,
		natGatewayAddressesResource,
		peerRoleARNResource,
		cpRouteTablesResource,
		vpcCIDRResource,
		tccpVPCIDResource,
		tccpOutputsResource,
		tccpSubnetsResource,
		regionResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		ipamResource,
		bridgeZoneResource,
		tccpEncryptionResource,
		s3BucketResource,
		s3ObjectResource,
		tccpAZsResource,
		tccpiResource,
		tccpResource,
		tccpfResource,
		serviceResource,
		endpointsResource,
		secretFinalizerResource,

		// All these resources implement cleanup functionality only being executed
		// on delete events.
		cleanupEBSVolumesResource,
		cleanupLoadBalancersResource,
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
		cr, err := key.ToCluster(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
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
