package ipam

import (
	"context"
	"fmt"
	"math/bits"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/key"
)

func init() {
	// Seed RNG for AZ shuffling.
	rand.Seed(time.Now().UnixNano())
}

// EnsureCreated allocates guest cluster network segment. It gathers existing
// subnets from existing AWSConfig/Status objects and existing VPCs from AWS.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error

	var cr v1alpha1.AWSConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "fetching latest version of custom resource")

		oldObj, err := key.ToCustomObject(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		newObj, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(oldObj.GetNamespace()).Get(oldObj.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		cr = *newObj

		r.logger.LogCtx(ctx, "level", "debug", "message", "fetched latest version of custom resource")
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if subnet needs to be allocated for cluster")

	if key.StatusNetworkCIDR(cr) == "" {
		var statusAZs []v1alpha1.AWSConfigStatusAWSAvailabilityZone
		var subnetCIDR net.IPNet
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "allocating cluster subnet CIDR")

			subnetCIDR, err = r.allocateSubnet(ctx)
			if err != nil {
				return microerror.Mask(err)
			}

			randomAZs, err := r.selectRandomAZs(key.SpecAvailabilityZones(cr))
			if err != nil {
				return microerror.Mask(err)
			}

			statusAZs, err = splitSubnetToStatusAZs(subnetCIDR, randomAZs)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("allocated cluster subnet CIDR %#q", subnetCIDR))
		}

		// Once we have all information together, regardless the update path, we
		// update the CR status. Note that we try to use the latest resource version
		// of the CR to get the status update properly sorted without any conflict.
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR status")

			cr.Status.Cluster.Network.CIDR = subnetCIDR.String()
			cr.Status.AWS.AvailabilityZones = statusAZs

			_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(cr.Namespace).UpdateStatus(&cr)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "updated CR status")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
		}

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "found out subnet doesn't need to be allocated for cluster")
	}

	return nil
}

func (r *Resource) allocateSubnet(ctx context.Context) (net.IPNet, error) {
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding allocated subnets from AWSConfigs")

		subnets, err := getAWSConfigSubnets(r.g8sClient)
		if err != nil {
			return microerror.Mask(err)
		}
		mutex.Lock()
		reservedSubnets = append(reservedSubnets, subnets...)
		mutex.Unlock()

		r.logger.LogCtx(ctx, "level", "debug", "message", "found allocated subnets from AWSConfigs")

		return nil
	})

	err = g.Wait()
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	reservedSubnets = ipam.CanonicalizeSubnets(r.networkRange, reservedSubnets)

	var subnet net.IPNet
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")

		subnet, err = ipam.Free(r.networkRange, r.allocatedSubnetMask, reservedSubnets)
		if err != nil {
			return net.IPNet{}, microerror.Maskf(err, "networkRange: %s, allocatedSubnetMask: %s, reservedSubnets: %#v", r.networkRange.String(), r.allocatedSubnetMask.String(), reservedSubnets)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet %#q", subnet.String()))
	}

	return subnet, nil
}

func (r *Resource) selectRandomAZs(n int) ([]string, error) {
	if n > len(r.availabilityZones) {
		return nil, microerror.Maskf(invalidParameterError, "requested nubmer of AZs %d is bigger than number of available AZs %d", n, len(r.availabilityZones))
	}

	// availabilityZones must be copied so that original slice doesn't get shuffled.
	shuffledAZs := make([]string, len(r.availabilityZones))
	copy(shuffledAZs, r.availabilityZones)
	rand.Shuffle(len(shuffledAZs), func(i, j int) {
		shuffledAZs[i], shuffledAZs[j] = shuffledAZs[j], shuffledAZs[i]
	})

	shuffledAZs = shuffledAZs[0:n]
	sort.Strings(shuffledAZs)
	return shuffledAZs, nil
}

func getAWSConfigSubnets(g8sClient versioned.Interface) ([]net.IPNet, error) {
	awsConfigList, err := g8sClient.ProviderV1alpha1().AWSConfigs(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, ac := range awsConfigList.Items {
		cidr := key.StatusNetworkCIDR(ac)
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

// splitSubnetToStatusAZs splits subnet such that each AZ gets private and
// public network. Size of these subnets depends on subnet.Mask and number of
// AZs.
func splitSubnetToStatusAZs(subnet net.IPNet, AZs []string) ([]v1alpha1.AWSConfigStatusAWSAvailabilityZone, error) {
	subnets, err := splitNetwork(subnet, uint(len(AZs)*2))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var statusAZs []v1alpha1.AWSConfigStatusAWSAvailabilityZone
	subnetIdx := 0
	for _, az := range AZs {
		private := subnets[subnetIdx]
		subnetIdx++
		public := subnets[subnetIdx]
		subnetIdx++

		statusAZ := v1alpha1.AWSConfigStatusAWSAvailabilityZone{
			Name: az,
			Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
				Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
					CIDR: private.String(),
				},
				Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
					CIDR: public.String(),
				},
			},
		}

		statusAZs = append(statusAZs, statusAZ)
	}

	return statusAZs, nil
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
