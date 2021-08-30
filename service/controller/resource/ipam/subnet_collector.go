package ipam

import (
	"context"
	"net"
	"reflect"
	"sync"

	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type SubnetCollectorConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	NetworkRange net.IPNet
}

type SubnetCollector struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	networkRange net.IPNet
}

func NewSubnetCollector(config SubnetCollectorConfig) (*SubnetCollector, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if reflect.DeepEqual(config.NetworkRange, net.IPNet{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkRange must not be empty", config)
	}

	c := &SubnetCollector{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		networkRange: config.NetworkRange,
	}

	return c, nil
}

func (c *SubnetCollector) Collect(ctx context.Context, networkRange net.IPNet) ([]net.IPNet, error) {
	var err error
	var mutex sync.Mutex
	var reservedSubnets []net.IPNet

	g := &errgroup.Group{}

	g.Go(func() error {
		c.logger.Debugf(ctx, "finding allocated subnets from AWSConfig CRs")

		subnets, err := c.getSubnetsFromAWSConfigs(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		c.logger.Debugf(ctx, "found allocated subnets from AWSConfig CRs")

		return nil
	})

	g.Go(func() error {
		c.logger.Debugf(ctx, "finding allocated subnets from Cluster CRs")

		subnets, err := c.getSubnetsFromClusters(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		c.logger.Debugf(ctx, "found allocated subnets from Cluster CRs")

		return nil
	})

	g.Go(func() error {
		c.logger.Debugf(ctx, "finding allocated subnets from MachineDeployment CRs")

		subnets, err := c.getSubnetsFromMachineDeployments(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		c.logger.Debugf(ctx, "found allocated subnets from MachineDeployment CRs")

		return nil
	})

	g.Go(func() error {
		c.logger.Debugf(ctx, "finding allocated subnets from VPCs")

		subnets, err := c.getSubnetsFromVPCs(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		c.logger.Debugf(ctx, "found allocated subnets from VPCs")

		return nil
	})

	err = g.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Here we decide which network range to consider when collected allocated
	// subnets. The given network range is the custom network range configured
	// in the NetworkPool CR. If it is empty we simply fall back to the network
	// range configured in the control plane.
	var nr net.IPNet
	{
		nr = networkRange
		if nr.IP.Equal(net.IP{}) {
			nr = c.networkRange
		}
	}

	reservedSubnets = ipam.CanonicalizeSubnets(nr, reservedSubnets)

	return reservedSubnets, nil
}

func (c *SubnetCollector) getSubnetsFromAWSConfigs(ctx context.Context) ([]net.IPNet, error) {
	awsConfigList, err := c.g8sClient.ProviderV1alpha1().AWSConfigs(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, ac := range awsConfigList.Items {
		cidr := key.StatusAWSConfigNetworkCIDR(ac)
		if cidr == "" {
			continue
		}

		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		results = append(results, *n)
	}

	return results, nil
}

func (c *SubnetCollector) getSubnetsFromClusters(ctx context.Context) ([]net.IPNet, error) {
	clusterList, err := c.g8sClient.InfrastructureV1alpha3().AWSClusters(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, c := range clusterList.Items {
		cidr := key.StatusClusterNetworkCIDR(c)
		if cidr == "" {
			continue
		}

		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		results = append(results, *n)
	}

	return results, nil
}

func (c *SubnetCollector) getSubnetsFromMachineDeployments(ctx context.Context) ([]net.IPNet, error) {
	machineDeploymentList, err := c.g8sClient.InfrastructureV1alpha3().AWSMachineDeployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, md := range machineDeploymentList.Items {
		cidr := key.MachineDeploymentSubnet(md)
		if cidr == "" {
			continue
		}

		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		results = append(results, *n)
	}

	return results, nil
}

func (c *SubnetCollector) getSubnetsFromVPCs(ctx context.Context) ([]net.IPNet, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	out, err := cc.Client.TenantCluster.AWS.EC2.DescribeSubnets(nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, subnet := range out.Subnets {
		_, n, err := net.ParseCIDR(*subnet.CidrBlock)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		results = append(results, *n)
	}

	return results, nil
}
