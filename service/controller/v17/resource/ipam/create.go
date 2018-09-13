package ipam

import (
	"context"
	"fmt"
	"net"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/v17/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v17/key"
)

// EnsureCreated allocates guest cluster network segment. It gathers existing
// subnets from existing AWSConfig/Status objects and existing VPCs from AWS.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "checking if subnet needs to be allocated for cluster")

	if key.ClusterNetworkCIDR(customObject) == "" {
		var subnetCIDR string
		// TODO(tuommaki): Remove this when all tenant clusters are upgraded to
		// this version and have subnet allocation in their Status field.
		// Tracked here: https://github.com/giantswarm/giantswarm/issues/4192.
		if key.CIDR(customObject) != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "using cluster CIDR from legacy field in CR")

			subnetCIDR = key.CIDR(customObject)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "allocating subnet for cluster")

			subnetCIDR, err = r.allocateSubnet(ctx)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Ensure that latest version of customObject is used.
		customObject, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(customObject.Namespace).Get(customObject.Name, apismetav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Status.Cluster.Network.CIDR = subnetCIDR
		_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(customObject.Namespace).UpdateStatus(customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("subnet %s allocated for cluster", subnetCIDR))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "subnet doesn't need to be allocated for cluster")
	}

	return nil
}

func (r *Resource) allocateSubnet(ctx context.Context) (string, error) {
	var reservedSubnets []net.IPNet
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting allocated subnets from VPCs")
		vpcSubnets, err := getVPCSubnets(ctx)
		if err != nil {
			return "", microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got allocated subnets from VPCs")

		reservedSubnets = append(reservedSubnets, vpcSubnets...)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting allocated subnets from AWSConfigs")
		awsConfigSubnets, err := getAWSConfigSubnets(r.g8sClient)
		if err != nil {
			return "", microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got allocated subnets from AWSConfigs")

		reservedSubnets = append(reservedSubnets, awsConfigSubnets...)
	}

	{
		reservedSubnets = canonicalizeSubnets(r.networkRange, reservedSubnets)

		r.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")
		subnet, err := ipam.Free(r.networkRange, r.allocatedSubnetMask, reservedSubnets)
		if err != nil {
			return "", microerror.Maskf(err, "networkRange: %s, allocatedSubnetMask: %s, reservedSubnets: %#v", r.networkRange.String(), r.allocatedSubnetMask.String(), reservedSubnets)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet: %s", subnet.String()))
		return subnet.String(), nil
	}
}

func canonicalizeSubnets(network net.IPNet, subnets []net.IPNet) []net.IPNet {
	// Naive deduplication as net.IPNet cannot be used as key for map. This
	// should be ok for current foreseeable future.
	for i := 0; i < len(subnets); i++ {
		// Remove subnets that don't belong to our desired network.
		if !network.Contains(subnets[i].IP) {
			subnets = append(subnets[:i], subnets[i+1:]...)
			i--
			continue
		}

		// Remove duplicates.
		for j := i + 1; j < len(subnets); j++ {
			if reflect.DeepEqual(subnets[i], subnets[j]) {
				subnets = append(subnets[:j], subnets[j+1:]...)
				j--
			}
		}
	}

	return subnets
}

func getAWSConfigSubnets(g8sClient versioned.Interface) ([]net.IPNet, error) {
	awsConfigList, err := g8sClient.ProviderV1alpha1().AWSConfigs(apismetav1.NamespaceAll).List(apismetav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, ac := range awsConfigList.Items {
		cidr := key.ClusterNetworkCIDR(ac)
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
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	subsOut, err := sc.AWSClient.EC2.DescribeSubnets(nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var results []net.IPNet
	for _, subnet := range subsOut.Subnets {
		_, n, err := net.ParseCIDR(*subnet.CidrBlock)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		results = append(results, *n)
	}

	return results, nil
}
