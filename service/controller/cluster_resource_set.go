package controller

import (
	"context"
	"fmt"
	"net"
	"strings"

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
	"github.com/giantswarm/tenantcluster"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/apiendpoint"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/bridgezone"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanupebsvolumes"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanuploadbalancers"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanupmachinedeployments"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanuprecordsets"
	"github.com/giantswarm/aws-operator/service/controller/resource/cleanupsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/cproutetables"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpvpc"
	"github.com/giantswarm/aws-operator/service/controller/resource/encryptionensurer"
	"github.com/giantswarm/aws-operator/service/controller/resource/endpoints"
	"github.com/giantswarm/aws-operator/service/controller/resource/eniconfigcrs"
	"github.com/giantswarm/aws-operator/service/controller/resource/ensurecpcrs"
	"github.com/giantswarm/aws-operator/service/controller/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/resource/natgatewayaddresses"
	"github.com/giantswarm/aws-operator/service/controller/resource/peerrolearn"
	"github.com/giantswarm/aws-operator/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3bucket"
	"github.com/giantswarm/aws-operator/service/controller/resource/secretfinalizer"
	"github.com/giantswarm/aws-operator/service/controller/resource/service"
	"github.com/giantswarm/aws-operator/service/controller/resource/snapshotid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpf"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpi"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcidstatus"
	"github.com/giantswarm/aws-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/service/internal/locker"
)

type clusterResourceSetConfig struct {
	CertsSearcher      certs.Interface
	HostAWSConfig      aws.Config
	K8sClient          k8sclient.Interface
	Locker             locker.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               tccp.ConfigAPIWhitelist
	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DockerDaemonCIDR           string
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
	ClusterDomain              string
	NetworkSetupDockerImage    string
	Route53Enabled             bool
	RouteTables                string
	PodInfraContainerImage     string
	RegistryDomain             string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
}

func newClusterResourceSet(config clusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var encrypterObject encrypter.Interface
	{
		c := &kms.EncrypterConfig{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		encrypterObject, err = kms.NewEncrypter(c)
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

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: config.CertsSearcher,
			Logger:        config.Logger,

			// TODO use a dedicated aws-operator key-pair.
			//
			//     https://github.com/giantswarm/giantswarm/issues/9327
			//
			CertID: certs.ClusterOperatorAPICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterChecker *ipam.ClusterChecker
	{
		c := ipam.ClusterCheckerConfig{
			G8sClient: config.K8sClient.G8sClient(),
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
			G8sClient: config.K8sClient.G8sClient(),
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
			G8sClient: config.K8sClient.G8sClient(),
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

	var apiEndpointResource resource.Interface
	{
		c := apiendpoint.Config{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		apiEndpointResource, err = apiendpoint.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsClientResource resource.Interface
	{
		c := awsclient.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var snapshotIDResource resource.Interface
	{
		c := snapshotid.Config{
			Logger: config.Logger,
		}

		snapshotIDResource, err = snapshotid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpAZsResource resource.Interface
	{
		c := tccpazs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			CIDRBlockAWSCNI: fmt.Sprintf("%s/%d", config.CalicoSubnet, config.CalicoCIDR),
		}

		tccpAZsResource, err = tccpazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionEnsurerResource resource.Interface
	{
		c := encryptionensurer.Config{
			Encrypter: encrypterObject,
			Logger:    config.Logger,
		}

		encryptionEnsurerResource, err = encryptionensurer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSecurityGroupsResource resource.Interface
	{
		c := tccpsecuritygroups.Config{
			ToClusterFunc: key.ToCluster,
			Logger:        config.Logger,
		}

		tccpSecurityGroupsResource, err = tccpsecuritygroups.New(c)
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

	var cleanupMachineDeploymentsResource resource.Interface
	{
		c := cleanupmachinedeployments.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		cleanupMachineDeploymentsResource, err = cleanupmachinedeployments.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cleanupRecordSets resource.Interface
	{
		c := cleanuprecordsets.Config{
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		cleanupRecordSets, err = cleanuprecordsets.New(c)
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
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,

			APIWhitelist:       config.APIWhitelist,
			CIDRBlockAWSCNI:    fmt.Sprintf("%s/%d", config.CalicoSubnet, config.CalicoCIDR),
			Detection:          tccpChangeDetection,
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
			ToClusterFunc:  key.ToCluster,
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

	var tccpVPCIDStatusResource resource.Interface
	{
		c := tccpvpcidstatus.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		tccpVPCIDStatusResource, err = tccpvpcidstatus.New(c)
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

	var eniConfigCRsResource resource.Interface
	{
		c := eniconfigcrs.Config{
			Logger: config.Logger,
		}

		eniConfigCRsResource, err = eniconfigcrs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ensureCPCRsResource resource.Interface
	{
		c := ensurecpcrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ensureCPCRsResource, err = ensurecpcrs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpVPCResource resource.Interface
	{
		c := cpvpc.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		cpVPCResource, err = cpvpc.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger: config.Logger,
			Tenant: tenantCluster,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// All these resources only fetch information from remote APIs and put them
		// into the controller context.
		awsClientResource,
		snapshotIDResource,
		accountIDResource,
		natGatewayAddressesResource,
		peerRoleARNResource,
		cpRouteTablesResource,
		cpVPCResource,
		tccpVPCIDResource,
		tccpOutputsResource,
		tccpSubnetsResource,
		regionResource,
		tenantClientsResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		encryptionEnsurerResource,
		apiEndpointResource,
		ipamResource,
		bridgeZoneResource,
		tccpSecurityGroupsResource,
		s3BucketResource,
		tccpAZsResource,
		tccpiResource,
		tccpResource,
		tccpfResource,
		serviceResource,
		endpointsResource,
		eniConfigCRsResource,
		ensureCPCRsResource,
		secretFinalizerResource,

		// All these resources implement logic to update CR status information.
		tccpVPCIDStatusResource,

		// All these resources implement cleanup functionality only being executed
		// on delete events.
		cleanupEBSVolumesResource,
		cleanupLoadBalancersResource,
		cleanupMachineDeploymentsResource,
		cleanupRecordSets,
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

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
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

func toCRUDResource(logger micrologger.Logger, ops crud.Interface) (*crud.Resource, error) {
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
