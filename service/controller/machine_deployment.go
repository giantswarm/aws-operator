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
	"github.com/giantswarm/operatorkit/v8/pkg/controller"
	"github.com/giantswarm/operatorkit/v8/pkg/resource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys/v3"
	"github.com/giantswarm/tenantcluster/v6/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v15/client/aws"
	"github.com/giantswarm/aws-operator/v15/pkg/label"
	"github.com/giantswarm/aws-operator/v15/pkg/project"
	"github.com/giantswarm/aws-operator/v15/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v15/service/controller/key"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/asgname"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/asgstatus"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/cleanuptcnpiamroles"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/cproutetables"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/cpvpc"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/ipam"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpnatgateways"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpvpcpcx"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnp"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnpazs"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnpf"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnpinstanceinfo"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnpoutputs"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnpsecuritygroups"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tcnpstatus"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tenantclients"
	"github.com/giantswarm/aws-operator/v15/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v15/service/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/v15/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v15/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v15/service/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/v15/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/v15/service/internal/images"
	"github.com/giantswarm/aws-operator/v15/service/internal/locker"
	event "github.com/giantswarm/aws-operator/v15/service/internal/recorder"
	"github.com/giantswarm/aws-operator/v15/service/internal/releases"
)

type MachineDeploymentConfig struct {
	CertsSearcher      certs.Interface
	CloudTags          cloudtags.Interface
	Event              event.Interface
	HAMaster           hamaster.Interface
	Images             images.Interface
	K8sClient          k8sclient.Interface
	Locker             locker.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	AlikeInstances             string
	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DockerDaemonCIDR           string
	DockerhubToken             string
	ExternalSNAT               bool
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	HostAWSConfig              aws.Config
	IgnitionPath               string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	ClusterDomain              string
	NetworkSetupDockerImage    string
	PodInfraContainerImage     string
	RegistryDomain             string
	RegistryMirrors            []string
	RouteTables                string
	SSHUserList                string
	SSOPublicKey               string
}

type MachineDeployment struct {
	*controller.Controller
}

func NewMachineDeployment(config MachineDeploymentConfig) (*MachineDeployment, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.DockerhubToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DockerhubToken must not be empty", config)
	}

	var err error

	var resources []resource.Interface
	{
		resources, err = newMachineDeploymentResources(config)
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
				return new(infrastructurev1alpha3.AWSMachineDeployment)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-machine-deployment-controller.
			Name: project.Name() + "-machine-deployment-controller",
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &MachineDeployment{
		Controller: operatorkitController,
	}

	return c, nil
}

func newMachineDeploymentResources(config MachineDeploymentConfig) ([]resource.Interface, error) {
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

	var cloudtagObject cloudtags.Interface
	{
		c := cloudtags.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		cloudtagObject, err = cloudtags.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomKeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomKeysSearcher, err = randomkeys.NewSearcher(c)
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

	var machineDeploymentChecker *ipam.MachineDeploymentChecker
	{
		c := ipam.MachineDeploymentCheckerConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		machineDeploymentChecker, err = ipam.NewMachineDeploymentChecker(c)
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

	var rel releases.Interface
	{
		c := releases.Config{
			K8sClient: config.K8sClient,
		}

		rel, err = releases.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpChangeDetection *changedetection.TCNP
	{
		c := changedetection.TCNPConfig{
			Logger:   config.Logger,
			Event:    config.Event,
			Releases: rel,
		}

		tcnpChangeDetection, err = changedetection.NewTCNP(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpCloudConfig *cloudconfig.TCNP
	{
		c := cloudconfig.TCNPConfig{
			Config: cloudconfig.Config{
				CertsSearcher:      certsSearcher,
				Encrypter:          encrypterObject,
				Event:              config.Event,
				HAMaster:           config.HAMaster,
				Images:             config.Images,
				K8sClient:          config.K8sClient,
				Logger:             config.Logger,
				RandomKeysSearcher: randomKeysSearcher,

				CalicoCIDR:              config.CalicoCIDR,
				CalicoMTU:               config.CalicoMTU,
				CalicoSubnet:            config.CalicoSubnet,
				ClusterIPRange:          config.ClusterIPRange,
				DockerDaemonCIDR:        config.DockerDaemonCIDR,
				DockerhubToken:          config.DockerhubToken,
				ExternalSNAT:            config.ExternalSNAT,
				IgnitionPath:            config.IgnitionPath,
				ClusterDomain:           config.ClusterDomain,
				NetworkSetupDockerImage: config.NetworkSetupDockerImage,
				PodInfraContainerImage:  config.PodInfraContainerImage,
				RegistryDomain:          config.RegistryDomain,
				RegistryMirrors:         config.RegistryMirrors,
				SSHUserList:             config.SSHUserList,
				SSOPublicKey:            config.SSOPublicKey,
			},
		}

		tcnpCloudConfig, err = cloudconfig.NewTCNP(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentPersister *ipam.MachineDeploymentPersister
	{
		c := ipam.MachineDeploymentPersisterConfig{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		machineDeploymentPersister, err = ipam.NewMachineDeploymentPersister(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,
			CertID:        certs.AWSOperatorAPICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger: config.Logger,
			Tenant: tenantCluster,

			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tenantClientsResource, err = tenantclients.New(c)
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
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var asgNameResource resource.Interface
	{
		c := asgname.Config{
			Logger: config.Logger,

			Stack:        key.StackTCNP,
			TagKey:       key.TagMachineDeployment,
			TagValueFunc: key.MachineDeploymentID,
		}

		asgNameResource, err = asgname.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var asgStatusResource resource.Interface
	{
		c := asgstatus.Config{
			Logger: config.Logger,
		}

		asgStatusResource, err = asgstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cleanupIAMRolesResource resource.Interface
	{
		c := cleanuptcnpiamroles.Config{
			Logger: config.Logger,
		}

		cleanupIAMRolesResource, err = cleanuptcnpiamroles.New(c)
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

	var ipamResource resource.Interface
	{
		c := ipam.Config{
			Checker:   machineDeploymentChecker,
			Collector: subnetCollector,
			K8sClient: config.K8sClient,
			Locker:    config.Locker,
			Logger:    config.Logger,
			Persister: machineDeploymentPersister,

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

	var s3ObjectResource resource.Interface
	{
		c := s3object.Config{
			CloudConfig: tcnpCloudConfig,
			Encrypter:   encrypterObject,
			Logger:      config.Logger,
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

	var tcnpAZsResource resource.Interface
	{
		c := tcnpazs.Config{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		tcnpAZsResource, err = tcnpazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpNATGatewaysResource resource.Interface
	{
		c := tccpnatgateways.Config{
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tccpNATGatewaysResource, err = tccpnatgateways.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var regionResource resource.Interface
	{
		c := region.Config{
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		regionResource, err = region.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpVPCPCXResource resource.Interface
	{
		c := tccpvpcpcx.Config{
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tccpVPCPCXResource, err = tccpvpcpcx.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSecurityGroupsResource resource.Interface
	{
		c := tccpsecuritygroups.Config{
			CtrlClient:    config.K8sClient.CtrlClient(),
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tccpSecurityGroupsResource, err = tccpsecuritygroups.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpSecurityGroupsResource resource.Interface
	{
		c := tcnpsecuritygroups.Config{
			Logger: config.Logger,
		}

		tcnpSecurityGroupsResource, err = tcnpsecuritygroups.New(c)
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

	var tcnpResource resource.Interface
	{
		c := tcnp.Config{
			CloudTags: cloudtagObject,
			Detection: tcnpChangeDetection,
			Encrypter: encrypterObject,
			Event:     config.Event,
			Images:    config.Images,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			AlikeInstances:   config.AlikeInstances,
			InstallationName: config.InstallationName,
		}

		tcnpResource, err = tcnp.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpfResource resource.Interface
	{
		c := tcnpf.Config{
			Event:  config.Event,
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		tcnpfResource, err = tcnpf.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpInstanceInfoResource resource.Interface
	{
		c := tcnpinstanceinfo.Config{
			Logger: config.Logger,
		}

		tcnpInstanceInfoResource, err = tcnpinstanceinfo.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpStatusResource resource.Interface
	{
		c := tcnpstatus.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		tcnpStatusResource, err = tcnpstatus.New(c)
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

	var tccpVPCIDResource resource.Interface
	{
		c := tccpvpcid.Config{
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tccpVPCIDResource, err = tccpvpcid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpOutputsResource resource.Interface
	{
		c := tcnpoutputs.Config{
			Logger: config.Logger,
		}

		tcnpOutputsResource, err = tcnpoutputs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// All these resources only fetch information from remote APIs and put them
		// into the controller context.
		awsClientResource,
		tenantClientsResource,
		accountIDResource,
		regionResource,
		cpRouteTablesResource,
		cpVPCResource,
		tccpNATGatewaysResource,
		tccpSecurityGroupsResource,
		tccpVPCIDResource,
		tccpVPCPCXResource,
		tccpSubnetsResource,
		tccpAZsResource,
		asgNameResource,
		asgStatusResource,
		tcnpAZsResource,
		tcnpOutputsResource,
		tcnpInstanceInfoResource,
		tcnpSecurityGroupsResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		s3ObjectResource,
		ipamResource,
		tcnpResource,
		tcnpfResource,

		// All these resources implement logic to update CR status information.
		tcnpStatusResource,

		// All these resources implement cleanup functionality only being executed
		// on delete events.
		cleanupIAMRolesResource,
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

func newMachineDeploymentToClusterFunc(ctrlClient ctrlClient.Client) func(ctx context.Context, obj interface{}) (infrastructurev1alpha3.AWSCluster, error) {
	return func(ctx context.Context, obj interface{}) (infrastructurev1alpha3.AWSCluster, error) {
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return infrastructurev1alpha3.AWSCluster{}, microerror.Mask(err)
		}

		m := &infrastructurev1alpha3.AWSCluster{}
		err = ctrlClient.Get(ctx, client.ObjectKey{Name: key.ClusterID(&cr), Namespace: cr.Namespace}, m)
		if err != nil {
			return infrastructurev1alpha3.AWSCluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
