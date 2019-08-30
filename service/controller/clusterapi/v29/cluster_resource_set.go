package v29

import (
	"context"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/bridgezone"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cleanupebsvolumes"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cleanuploadbalancers"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cleanupsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cproutetables"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cpvpccidr"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/encryption"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/natgatewayaddresses"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/peerrolearn"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/secretfinalizer"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/service"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccp"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpf"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpi"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccproutetables"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpsubnets"
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

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
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
			OIDC:                      config.OIDC,
			PodInfraContainerImage:    config.PodInfraContainerImage,
			RegistryDomain:            config.RegistryDomain,
			SSHUserList:               config.SSHUserList,
			SSOPublicKey:              config.SSOPublicKey,
		}

		cloudConfig, err = cloudconfig.New(c)
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

	var awsClientResource controller.Resource
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

	var tccpAZsResource controller.Resource
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

	var encryptionResource controller.Resource
	{
		c := encryption.Config{
			Encrypter:     encrypterObject,
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,
		}

		encryptionResource, err = encryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ipamResource controller.Resource
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

	var cleanupEBSVolumesResource controller.Resource
	{
		c := cleanupebsvolumes.Config{
			Logger: config.Logger,
		}

		cleanupEBSVolumesResource, err = cleanupebsvolumes.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cleanupLoadBalancersResource controller.Resource
	{
		c := cleanuploadbalancers.Config{
			Logger: config.Logger,
		}

		cleanupLoadBalancersResource, err = cleanuploadbalancers.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cleanupSecurityGroups controller.Resource
	{
		c := cleanupsecuritygroups.Config{
			Logger: config.Logger,
		}

		cleanupSecurityGroups, err = cleanupsecuritygroups.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var regionResource controller.Resource
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

	var tccpResource controller.Resource
	{
		c := tccp.Config{
			CMAClient:            config.CMAClient,
			EncrypterRoleManager: encrypterRoleManager,
			Logger:               config.Logger,

			APIWhitelist:       config.APIWhitelist,
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

	var tccpRouteTablesResource controller.Resource
	{
		c := tccproutetables.Config{
			Logger: config.Logger,
		}

		tccpRouteTablesResource, err = tccproutetables.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSubnetsResource controller.Resource
	{
		c := tccpsubnets.Config{
			Logger: config.Logger,
		}

		tccpSubnetsResource, err = tccpsubnets.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpfResource controller.Resource
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

	var tccpiResource controller.Resource
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

	var cpRouteTablesResource controller.Resource
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

	var secretFinalizerResource controller.Resource
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

	var vpcCIDRResource controller.Resource
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

	resources := []controller.Resource{
		// All these resources only fetch information from remote APIs and put them
		// into the controller context.
		awsClientResource,
		accountIDResource,
		natGatewayAddressesResource,
		peerRoleARNResource,
		cpRouteTablesResource,
		vpcCIDRResource,
		tccpOutputsResource,
		tccpRouteTablesResource,
		tccpSubnetsResource,
		regionResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		ipamResource,
		bridgeZoneResource,
		encryptionResource,
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
