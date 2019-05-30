package ipam

import (
	"context"
	"encoding/json"
	"fmt"
	"math/bits"
	"net"
	"sync"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
	"github.com/giantswarm/aws-operator/service/network"
)

// EnsureCreated allocates guest cluster network segment. It gathers existing
// subnets from existing AWSConfig/Status objects and existing VPCs from AWS.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var cr cmav1alpha1.Cluster
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "fetching latest version of custom resource")

		oldObj, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		newObj, err := r.cmaClient.ClusterV1alpha1().Clusters(oldObj.GetNamespace()).Get(oldObj.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		cr = *newObj

		r.logger.LogCtx(ctx, "level", "debug", "message", "fetched latest version of custom resource")
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if subnet needs to be allocated for cluster")

	if key.StatusClusterNetworkCIDR(cr) == "" {
		var subnetCIDR net.IPNet
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "allocating cluster subnet CIDR")

			callbacks := network.AllocationCallbacks{
				GetReservedNetworks:     r.getReservedNetworks,
				PersistAllocatedNetwork: r.persistAllocatedNetwork(cr, key.WorkerAvailabilityZones(cc.Status.TenantCluster.TCCP.MachineDeployment)),
			}

			subnetCIDR, err = r.networkAllocator.Allocate(ctx, r.networkRange, r.allocatedSubnetMask, callbacks)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated CR status with allocated cluster subnet CIDR %#q", subnetCIDR))

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)

		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "found out subnet doesn't need to be allocated for cluster")
	}

	return nil
}

func (r *Resource) getReservedNetworks(ctx context.Context) ([]net.IPNet, error) {
	var err error
	var mutex sync.Mutex
	var reservedSubnets []net.IPNet

	g := &errgroup.Group{}

	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding allocated subnets from VPCs")

		subnets, err := getVPCSubnets(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		r.logger.LogCtx(ctx, "level", "debug", "message", "found allocated subnets from VPCs")

		return nil
	})

	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding allocated subnets from AWSConfig CRs")

		subnets, err := getAWSConfigSubnets(r.g8sClient)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		r.logger.LogCtx(ctx, "level", "debug", "message", "found allocated subnets from AWSConfig CRs")

		return nil
	})

	g.Go(func() error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding allocated subnets from Cluster CRs")

		subnets, err := getClusterSubnets(r.cmaClient)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		r.logger.LogCtx(ctx, "level", "debug", "message", "found allocated subnets from Cluster CRs")

		return nil
	})

	err = g.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	reservedSubnets = ipam.CanonicalizeSubnets(r.networkRange, reservedSubnets)

	return reservedSubnets, nil
}

func (r *Resource) persistAllocatedNetwork(cr cmav1alpha1.Cluster, azs []string) func(ctx context.Context, subnet net.IPNet) error {
	return func(ctx context.Context, subnet net.IPNet) error {
		return r.splitAndPersistReservedSubnet(ctx, cr, subnet, azs)
	}
}

func (r *Resource) splitAndPersistReservedSubnet(ctx context.Context, cr cmav1alpha1.Cluster, subnet net.IPNet, azs []string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR status to persist network allocation and chosen availability zones")

	var providerStatus g8sv1alpha1.AWSClusterStatus

	err := json.Unmarshal(cr.Status.ProviderStatus.Raw, &providerStatus)
	if err != nil {
		return microerror.Mask(err)
	}

	providerStatus.Provider.Network.CIDR = subnet.String()

	b, err := json.Marshal(providerStatus)
	if err != nil {
		return microerror.Mask(err)
	}

	cr.Status.ProviderStatus.Raw = b

	_, err = r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).UpdateStatus(&cr)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "updated CR status to persist network allocation and chosen availability zones")

	return nil
}

func getAWSConfigSubnets(g8sClient versioned.Interface) ([]net.IPNet, error) {
	awsConfigList, err := g8sClient.ProviderV1alpha1().AWSConfigs(metav1.NamespaceAll).List(metav1.ListOptions{})
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

func getClusterSubnets(cmaClient clientset.Interface) ([]net.IPNet, error) {
	clusterList, err := cmaClient.Cluster().Clusters(metav1.NamespaceAll).List(metav1.ListOptions{})
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

func getVPCSubnets(ctx context.Context) ([]net.IPNet, error) {
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

// calculateSubnetMask calculates new subnet mask to accommodate n subnets.
func calculateSubnetMask(networkMask net.IPMask, n uint) (net.IPMask, error) {
	if n == 0 {
		return nil, microerror.Maskf(invalidParameterError, "divide by zero")
	}

	// Amount of bits needed to accommodate enough subnets for public and
	// private subnet in each AZ.
	subnetBitsNeeded := bits.Len(n - 1)

	maskOnes, maskBits := networkMask.Size()
	if subnetBitsNeeded > maskBits-maskOnes {
		return nil, microerror.Maskf(invalidParameterError, "no room in network mask %s to accommodate %d subnets", networkMask.String(), n)
	}

	return net.CIDRMask(maskOnes+subnetBitsNeeded, maskBits), nil
}

// splitNetwork returns n subnets from network.
func splitNetwork(network net.IPNet, n uint) ([]net.IPNet, error) {
	mask, err := calculateSubnetMask(network.Mask, n)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var subnets []net.IPNet
	for i := uint(0); i < n; i++ {
		subnet, err := ipam.Free(network, mask, subnets)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		subnets = append(subnets, subnet)
	}

	return subnets, nil
}
