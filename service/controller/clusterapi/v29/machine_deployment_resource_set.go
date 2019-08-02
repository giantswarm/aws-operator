package v29

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cpvpccidr"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/encryption"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpnatgateways"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tcnp"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tcnpazs"
)

func NewMachineDeploymentResourceSet(config MachineDeploymentResourceSetConfig) (*controller.ResourceSet, error) {
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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),

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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
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
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		encryptionResource, err = encryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ipamResource controller.Resource
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

	var tcnpAZsResource controller.Resource
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

	var tccpNATGatewaysResource controller.Resource
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

	var regionResource controller.Resource
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

	var tccpSecurityGroupsResource controller.Resource
	{
		c := tccpsecuritygroups.Config{
			Logger: config.Logger,
		}

		tccpSecurityGroupsResource, err = tccpsecuritygroups.New(c)
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

	var tcnpResource controller.Resource
	{
		c := tcnp.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,

			InstallationName: config.InstallationName,
		}

		tcnpResource, err = tcnp.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpVPCCIDRResource controller.Resource
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

	var tccpVPCIDResource controller.Resource
	{
		c := tccpvpcid.Config{
			CMAClient:     config.CMAClient,
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		tccpVPCIDResource, err = tccpvpcid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		awsClientResource,
		accountIDResource,
		tccpVPCIDResource,
		cpVPCCIDRResource,
		tccpNATGatewaysResource,
		tccpSecurityGroupsResource,
		tccpSubnetsResource,
		regionResource,
		encryptionResource,
		ipamResource,
		tccpAZsResource,
		tcnpAZsResource,
		tcnpResource,
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

		if key.OperatorVersion(&cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		cc := controllercontext.Context{}
		ctx = controllercontext.NewContext(ctx, cc)

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
