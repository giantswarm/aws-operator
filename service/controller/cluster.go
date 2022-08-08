package controller

import (
	"context"
	"fmt"
	"net"
	"strings"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v4/pkg/certs"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/crud"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v6/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/client/aws"
	"github.com/giantswarm/aws-operator/v13/pkg/label"
	"github.com/giantswarm/aws-operator/v13/pkg/project"
	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/apiendpoint"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/appsconfig"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/awscnicleaner"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/bridgezone"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanupebsvolumes"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanupenis"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanuploadbalancers"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanupmachinedeployments"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanuprecordsets"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanupsecuritygroups"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cleanupvpcpeerings"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cproutetables"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/cpvpc"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/encryptionensurer"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/endpoints"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/ipam"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/keepforcrs"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/natgatewayaddresses"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/peerrolearn"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/restrictawsnodedaemonset"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/s3bucket"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/secretfinalizer"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/service"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpf"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpi"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tccpvpcidstatus"
	"github.com/giantswarm/aws-operator/v13/service/controller/resource/tenantclients"
	"github.com/giantswarm/aws-operator/v13/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v13/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v13/service/internal/cphostedzone"
	"github.com/giantswarm/aws-operator/v13/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v13/service/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/v13/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/v13/service/internal/locker"
	event "github.com/giantswarm/aws-operator/v13/service/internal/recorder"
)

type ClusterConfig struct {
	CloudTags cloudtags.Interface
	Event     event.Interface
	K8sClient k8sclient.Interface
	HAMaster  hamaster.Interface
	Locker    locker.Interface
	Logger    micrologger.Logger

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               tccp.ConfigAPIWhitelist
	CalicoCIDR                 int
	CalicoSubnet               string
	DeleteLoggingBucket        bool
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	HostAWSConfig              aws.Config
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	RouteTables                string
	Route53Enabled             bool
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var resources []resource.Interface
	{
		resources, err = newClusterResources(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			InitCtx: func(ctx context.Context, obj interface{}) (context.Context, error) {
				return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() ctrlClient.Object {
				return new(infrastructurev1alpha3.AWSCluster)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-cluster-controller.
			Name: project.Name() + "-cluster-controller",
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: operatorkitController,
	}

	return c, nil
}

func newClusterResources(config ClusterConfig) ([]resource.Interface, error) {
	var err error

	var certsSearcher *certs.Searcher
	{
		c := certs.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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
			Event:  config.Event,
			Logger: config.Logger,
		}

		tccpChangeDetection, err = changedetection.NewTCCP(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpfChangeDetection *changedetection.TCCPF
	{
		c := changedetection.TCCPFConfig{
			Logger: config.Logger,
		}

		tccpfChangeDetection, err = changedetection.NewTCCPF(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
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
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		clusterChecker, err = ipam.NewClusterChecker(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var subnetCollector *ipam.SubnetCollector
	{
		c := ipam.SubnetCollectorConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,

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
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		clusterPersister, err = ipam.NewClusterPersister(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var hostedZone *cphostedzone.HostedZone
	{
		c := cphostedzone.Config{
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
		}

		hostedZone, err = cphostedzone.New(c)
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

	var keepForAWSControlPlaneCRsResource resource.Interface
	{
		c := keepforcrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewObjFunc: func() runtime.Object {
				return &infrastructurev1alpha3.AWSControlPlane{}
			},
		}

		keepForAWSControlPlaneCRsResource, err = keepforcrs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keepForAWSMachineDeploymentCRsResource resource.Interface
	{
		c := keepforcrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewObjFunc: func() runtime.Object {
				return &infrastructurev1alpha3.AWSMachineDeployment{}
			},
		}

		keepForAWSMachineDeploymentCRsResource, err = keepforcrs.New(c)
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

	var restrictAwsNodeDaemonsetResource resource.Interface
	{
		c := restrictawsnodedaemonset.Config{
			Logger: config.Logger,
		}

		restrictAwsNodeDaemonsetResource, err = restrictawsnodedaemonset.New(c)
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
			CtrlClient:    config.K8sClient.CtrlClient(),
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
			K8sClient: config.K8sClient,
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
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,

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

	var cleanupENIs resource.Interface
	{
		c := cleanupenis.Config{
			Logger: config.Logger,
		}

		cleanupENIs, err = cleanupenis.New(c)
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
			Event:      config.Event,
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
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

	var cleanupVPCPeerings resource.Interface
	{
		c := cleanupvpcpeerings.Config{
			Logger: config.Logger,
		}

		cleanupVPCPeerings, err = cleanupvpcpeerings.New(c)
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
			CloudTags:  config.CloudTags,
			Event:      config.Event,
			CtrlClient: config.K8sClient.CtrlClient(),
			HAMaster:   config.HAMaster,
			K8sClient:  config.K8sClient,
			Logger:     config.Logger,

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
			Detection:  tccpfChangeDetection,
			Event:      config.Event,
			HostedZone: hostedZone,
			Logger:     config.Logger,

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
			Event:  config.Event,
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
		var routeTableNames []string
		{
			if config.RouteTables != "" {
				routeTableNames = strings.Split(config.RouteTables, ",")
			}
		}

		c := cproutetables.Config{
			Logger:       config.Logger,
			Installation: config.InstallationName,

			Names: routeTableNames,
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

			ToClusterFunc: key.ToCluster,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appsConfigResource resource.Interface
	{
		c := appsconfig.Config{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		appsConfigResource, err = appsconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsCniCleanerResource resource.Interface
	{
		c := awscnicleaner.Config{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		awsCniCleanerResource, err = awscnicleaner.New(c)
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
		cpVPCResource,
		tccpVPCIDResource,
		tccpOutputsResource,
		tccpSubnetsResource,
		regionResource,
		tenantClientsResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		encryptionEnsurerResource,
		restrictAwsNodeDaemonsetResource,
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
		secretFinalizerResource,
		appsConfigResource,
		awsCniCleanerResource,

		// All these resources implement logic to update CR status information.
		tccpVPCIDStatusResource,

		// All these resources implement cleanup functionality only being executed
		// on delete events.
		cleanupEBSVolumesResource,
		cleanupLoadBalancersResource,
		cleanupMachineDeploymentsResource,
		cleanupRecordSets,
		cleanupSecurityGroups,
		cleanupENIs,
		cleanupVPCPeerings,
		keepForAWSControlPlaneCRsResource,
		keepForAWSMachineDeploymentCRsResource,
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

	return resources, nil
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
