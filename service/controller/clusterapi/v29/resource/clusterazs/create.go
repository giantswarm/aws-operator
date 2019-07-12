package clusterazs

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sort"
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
		// Collect non-empty subnets from AZ-subnet -pairs that belong to this
		// specific TCCP.
		for _, snetPair := range azs {
			if !reflect.DeepEqual(net.IPNet{}, snetPair.Public) {
				allocatedNetworks = append(allocatedNetworks, snetPair.Public)
			}

			if !reflect.DeepEqual(net.IPNet{}, snetPair.Private) {
				allocatedNetworks = append(allocatedNetworks, snetPair.Private)
			}
		}
	}

	{
		clusterCIDR := key.StatusClusterNetworkCIDR(cr)
		if clusterCIDR == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cluster network not yet allocated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		// Parse TCCP network CIDR.
		_, tccpSubnet, err := net.ParseCIDR(clusterCIDR)
		if err != nil {
			return microerror.Mask(err)
		}

		azs, err = r.ensureAZsAreAssignedWithSubnet(ctx, *tccpSubnet, azs)
		if err != nil {
			return microerror.Mask(err)
		}

		ccAZs := mapAZSubnetsToControllerContextTypes(azs)

		r.logger.LogCtx(ctx, "level", "debug", "message", azSubnetsToString(azs))

		cc.Status.TenantCluster.AvailabilityZones = ccAZs
	}

	return nil
}

// ensureAZsAreAssignedWithSubnet iterates over AZ-subnetPair map, removes
// subnets that are already in use from available subnets' list and then
// assigns one to AZs that don't have subnet assigned yet.
func (r *Resource) ensureAZsAreAssignedWithSubnet(ctx context.Context, tccpSubnet net.IPNet, azs map[string]subnetPair) (map[string]subnetPair, error) {
	// Split TCCP network between maximum number of AZs. This is because of
	// current limitation in IPAM design and AWS TCCP infrastructure
	// design.
	clusterAZSubnets, err := ipam.Split(tccpSubnet, key.MaximumNumberOfAZsInCluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var azNames []string
	{
		for az, _ := range azs {
			azNames = append(azNames, az)
		}

		sort.Strings(azNames)
	}

	// Remove already allocated networks from clusterAZSubnets before assigning
	// remaining subnets to AZs without themout them.
	for _, az := range azNames {
		subnets := azs[az]

		// Check if subnets of given availability zone already contain
		// value?
		if !subnets.areEmpty() {
			// Calculate the parent network from public subnet (always
			// present for functional AZ).
			parentNet := ipam.CalculateParent(subnets.Public)

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q already has subnet allocated: %q", az, parentNet.String()))

			// Filter out already allocated AZ subnet from available AZ
			// size networks.
			clusterAZSubnets = ipam.Filter(clusterAZSubnets, func(n net.IPNet) bool {
				return !reflect.DeepEqual(n, parentNet)
			})
		}
	}

	// Assign subnet to AZs that don't it yet.
	for _, az := range azNames {
		subnets := azs[az]

		// Only proceed with AZs that don't have subnet allocated yet.
		if subnets.areEmpty() {
			if len(clusterAZSubnets) > 0 {
				// Pick first available AZ subnet and split it to public
				// and private.
				ss, err := ipam.Split(clusterAZSubnets[0], 2)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q doesn't have subnet allocation - allocated: %q", az, clusterAZSubnets[0].String()))

				// Update available AZ subnets.
				clusterAZSubnets = clusterAZSubnets[1:]

				subnets.Public = ss[0]
				subnets.Private = ss[1]
				azs[az] = subnets
			} else {
				return nil, microerror.Maskf(invalidConfigError, "no more unallocated subnets left but there's this AZ still left: %q", az)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting cluster availability zones to controllercontext: %#v", azs))

	return azs, nil
}

// fromEC2SubnetsToMap extracts availability zones and public / private subnet
// CIDRs from given EC2 subnet slice and returns respectively structured map or
// error on invalid data.
func fromEC2SubnetsToMap(ss []*ec2.Subnet) (map[string]subnetPair, error) {
	azMap := make(map[string]subnetPair)

	for _, s := range ss {
		if s == nil || s.AvailabilityZone == nil || s.CidrBlock == nil || s.Tags == nil {
			return nil, microerror.Maskf(executionFailedError, "invalid subnet entry in controllercontext.Status.TenantCluster.TCCP.Subnets: %#v", s)
		}

		var subnetType string
		var subnetBelongsToTCCP bool
		for _, t := range s.Tags {
			if t == nil || t.Key == nil {
				return nil, microerror.Maskf(executionFailedError, "invalid tag in ec2.Subnet: %#v", s)
			}

			if t.Value == nil {
				// It's ok that tag doesn't have value. It's just not the one
				// we care about here.
				continue
			}

			switch *t.Key {
			case key.TagTCCP:
				subnetBelongsToTCCP = true
			case key.TagSubnetType:
				subnetType = strings.TrimSpace(*t.Value)
			}
		}

		if !subnetBelongsToTCCP {
			// VPC contains many subnets for various purposes in addition to
			// TCCP, mainly for node pools. We are only interested in TCCP
			// subnets in here.
			continue
		}

		mappedSubnet := azMap[*s.AvailabilityZone]

		_, cidr, err := net.ParseCIDR(*s.CidrBlock)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		switch subnetType {
		case "public":
			mappedSubnet.Public = *cidr
		case "private":
			mappedSubnet.Private = *cidr
		default:
			return nil, microerror.Maskf(invalidConfigError, "invalid subnet type in ec2.Subnet tag: %q: %q", key.TagSubnetType, subnetType)
		}

		azMap[*s.AvailabilityZone] = mappedSubnet
	}

	return azMap, nil
}

func mapAZSubnetsToControllerContextTypes(azs map[string]subnetPair) []controllercontext.ContextStatusTenantClusterAvailabilityZone {
	var results []controllercontext.ContextStatusTenantClusterAvailabilityZone

	for az, subnet := range azs {
		ccAZ := controllercontext.ContextStatusTenantClusterAvailabilityZone{
			Name:          az,
			PublicSubnet:  subnet.Public,
			PrivateSubnet: subnet.Private,
		}

		results = append(results, ccAZ)
	}

	return results
}

func azSubnetsToString(azs map[string]subnetPair) string {
	var result strings.Builder
	result.WriteString("availability zone subnet allocations: {")
	for az, subnet := range azs {
		result.WriteString(fmt.Sprintf("\n%q: [pub: %q, private: %q]", az, subnet.Public.String(), subnet.Private.String()))
	}

	if len(azs) > 0 {
		result.WriteString("\n}")
	} else {
		result.WriteString("}")
	}

	return result.String()
}
