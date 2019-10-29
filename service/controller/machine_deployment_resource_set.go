package controller

import (
	"context"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/asgstatus"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/cproutetables"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpvpccidr"
	"github.com/giantswarm/aws-operator/service/controller/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpnatgateways"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcpcx"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnp"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnpazs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnpencryption"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnpf"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnpoutputs"
	"github.com/giantswarm/aws-operator/service/versionbundle"
)

func newMachineDeploymentResourceSet(config machineDeploymentResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var encrypterObject encrypter.Interface
	{
		encrypterObject, err = newEncrypterObject(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentChecker *ipam.MachineDeploymentChecker
	{
		c := ipam.MachineDeploymentCheckerConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentChecker, err = ipam.NewMachineDeploymentChecker(c)
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

	var tcnpChangeDetection *changedetection.TCNP
	{
		c := changedetection.TCNPConfig{
			Logger: config.Logger,
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

		tcnpCloudConfig, err = cloudconfig.NewTCNP(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentPersister *ipam.MachineDeploymentPersister
	{
		c := ipam.MachineDeploymentPersisterConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentPersister, err = ipam.NewMachineDeploymentPersister(c)
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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
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

	var tccpAZsResource resource.Interface
	{
		c := tccpazs.Config{
			CMAClient:     config.CMAClient,
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		tccpAZsResource, err = tccpazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpEncryptionResource resource.Interface
	{
		c := tcnpencryption.Config{
			CMAClient: config.CMAClient,
			Encrypter: encrypterObject,
			Logger:    config.Logger,
		}

		tcnpEncryptionResource, err = tcnpencryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ipamResource resource.Interface
	{
		c := ipam.Config{
			Checker:   machineDeploymentChecker,
			Collector: subnetCollector,
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
			CertsSearcher:      config.CertsSearcher,
			CloudConfig:        tcnpCloudConfig,
			LabelsFunc:         key.KubeletLabelsTCNP,
			Logger:             config.Logger,
			CMAClient:          config.CMAClient,
			PathFunc:           key.S3ObjectPathTCNP,
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

	var tcnpAZsResource resource.Interface
	{
		c := tcnpazs.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		tccpVPCPCXResource, err = tccpvpcpcx.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSecurityGroupsResource resource.Interface
	{
		c := tccpsecuritygroups.Config{
			Logger: config.Logger,
		}

		tccpSecurityGroupsResource, err = tccpsecuritygroups.New(c)
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
			CMAClient: config.CMAClient,
			Detection: tcnpChangeDetection,
			Logger:    config.Logger,

			InstallationName: config.InstallationName,
		}

		tcnpResource, err = tcnp.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpASGStatusResource resource.Interface
	{
		c := asgstatus.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		tcnpASGStatusResource, err = asgstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpfResource resource.Interface
	{
		c := tcnpf.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		tcnpfResource, err = tcnpf.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpVPCCIDRResource resource.Interface
	{
		c := cpvpccidr.Config{
			Logger: config.Logger,

			VPCPeerID: config.VPCPeerID,
		}

		cpVPCCIDRResource, err = cpvpccidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpVPCIDResource resource.Interface
	{
		c := tccpvpcid.Config{
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
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
		accountIDResource,
		regionResource,
		cpRouteTablesResource,
		cpVPCCIDRResource,
		tccpNATGatewaysResource,
		tccpSecurityGroupsResource,
		tccpVPCIDResource,
		tccpVPCPCXResource,
		tccpSubnetsResource,
		tccpAZsResource,
		tcnpASGStatusResource,
		tcnpAZsResource,
		tcnpOutputsResource,
		tcnpEncryptionResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		s3ObjectResource,
		ipamResource,
		tcnpResource,
		tcnpfResource,
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
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == versionbundle.New().Version {
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

func newMachineDeploymentToClusterFunc(cmaClient clientset.Interface) func(obj interface{}) (v1alpha1.Cluster, error) {
	return func(obj interface{}) (v1alpha1.Cluster, error) {
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return v1alpha1.Cluster{}, microerror.Mask(err)
		}

		m, err := cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Get(key.ClusterID(&cr), metav1.GetOptions{})
		if err != nil {
			return v1alpha1.Cluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
