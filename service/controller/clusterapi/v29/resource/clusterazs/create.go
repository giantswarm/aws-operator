package clusterazs

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

// subnetPair is temporary type for mapping existing subnets from
// controllercontext to AZs.
type subnetPair struct {
	public  net.IPNet
	private net.IPNet
}

func (sp subnetPair) areEmpty() bool {
	return (sp.public.IP == nil && sp.public.Mask == nil) && (sp.private.IP == nil && sp.private.Mask == nil)

}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var machineDeployments []clusterv1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&v1.LabelSelector{},
			label.Cluster,
			key.ClusterID(&cr),
		)
		o := metav1.ListOptions{
			LabelSelector: labels.Set(l.MatchLabels).String(),
		}

		list, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		machineDeployments = list.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments)))
	}

	var azs map[string]subnetPair
	{
		// Acquire AZs together with corresponding subnets from AWS EC2 API
		// results.
		azs, err = fromEC2SubnetsToMap(cc.Status.TenantCluster.TCCP.Subnets)
		if err != nil {
			return microerror.Mask(err)
		}

		// Include master's AZ if missing.
		if _, exists := azs[key.MasterAvailabilityZone(cr)]; !exists {
			azs[key.MasterAvailabilityZone(cr)] = subnetPair{}
		}

		// Include worker AZs if missing.
		for _, md := range machineDeployments {
			for _, az := range key.WorkerAvailabilityZones(md) {
				if _, exists := azs[az]; !exists {
					azs[az] = subnetPair{}
				}
			}
		}
	}

	var allocatedNetworks []net.IPNet
	{
		// Collect non-empty subnets from AZ-subnet -pairs.
		for _, snetPair := range azs {
			if !reflect.DeepEqual(net.IPNet{}, snetPair.public) {
				allocatedNetworks = append(allocatedNetworks, snetPair.public)
			}

			if !reflect.DeepEqual(net.IPNet{}, snetPair.private) {
				allocatedNetworks = append(allocatedNetworks, snetPair.private)
			}
		}
	}

	{
		// Parse TCCP network CIDR.
		_, clusterCIDR, err := net.ParseCIDR(key.StatusClusterNetworkCIDR(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		// Split TCCP network between maximum number of AZs.
		clusterAZSubnets, err := ipam.Split(*clusterCIDR, key.MaximumNumberOfAZsInCluster)
		if err != nil {
			return microerror.Mask(err)
		}

		// Convert collected availability zones into controllercontext types
		// and allocate AZ level subnets when needed.
		var ccAZs []controllercontext.ContextStatusTenantClusterAvailabilityZone
		for az, subnets := range azs {
			ccAZ := controllercontext.ContextStatusTenantClusterAvailabilityZone{
				Name: az,
			}

			// Check if subnets of given availability zone already contain
			// value?
			if !subnets.areEmpty() {
				ccAZ.PublicSubnet = subnets.public
				ccAZ.PrivateSubnet = subnets.private

				// Calculate the parent network from public subnet (always
				// present for functional AZ).
				parentNet := ipam.CalculateParent(subnets.public)

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q already has subnet allocated: %q", az, parentNet.String()))

				// Filter out already allocated AZ subnet from available AZ
				// size networks.
				clusterAZSubnets = ipam.Filter(clusterAZSubnets, func(n net.IPNet) bool {
					return !reflect.DeepEqual(n, parentNet)
				})
			} else {
				if len(clusterAZSubnets) > 0 {
					// Pick first available AZ subnet and split it to public
					// and private.
					subnets, err := ipam.Split(clusterAZSubnets[0], 2)
					if err != nil {
						return microerror.Mask(err)
					}

					r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q doesn't have subnet allocation - allocated: %q", az, clusterAZSubnets[0].String()))

					// Update available AZ subnets.
					clusterAZSubnets = clusterAZSubnets[1:]

					// Persist allocated & split subnets.
					ccAZ.PublicSubnet = subnets[0]
					ccAZ.PrivateSubnet = subnets[1]
				} else {
					return microerror.Maskf(invalidConfigError, "no more unallocated subnets left but there's this AZ still left: %q", az)
				}
			}

			ccAZs = append(ccAZs, ccAZ)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting cluster availability zones to controllercontext: %#v", azs))

		cc.Status.TenantCluster.AvailabilityZones = ccAZs
	}

	return nil
}

// fromEC2SubnetsToMap extracts availability zones and public / private subnet
// CIDRs from given EC2 subnet slice and returns respectively structured map or
// error on invalid data.
func fromEC2SubnetsToMap(ss []*ec2.Subnet) (map[string]subnetPair, error) {
	azMap := make(map[string]subnetPair)

	for _, s := range ss {
		if s == nil || s.AvailabilityZone == nil || s.CidrBlock == nil || s.Tags == nil {
			continue
		}

		mappedSubnet := azMap[*s.AvailabilityZone]

		_, cidr, err := net.ParseCIDR(*s.CidrBlock)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var subnetType string
		for _, t := range s.Tags {
			if t == nil || t.Key == nil || t.Value == nil {
				continue
			}

			if *t.Key == key.TagEC2SubnetType {
				subnetType = strings.TrimSpace(*t.Value)
			}
		}

		switch subnetType {
		case "public":
			mappedSubnet.public = *cidr
		case "private":
			mappedSubnet.private = *cidr
		default:
			return nil, microerror.Maskf(invalidConfigError, "invalid subnet type in ec2.Subnet tag: %q: %q", key.TagEC2SubnetType, subnetType)
		}

		azMap[*s.AvailabilityZone] = mappedSubnet
	}

	return azMap, nil
}
