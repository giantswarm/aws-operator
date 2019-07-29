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
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// We need to cancel the resource early in case the ipam resource did not yet
	// allocate a subnet for the tenant cluster.
	if key.StatusClusterNetworkCIDR(cr) == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cannot collect private and public subnets for availability zones")
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster subnet not yet allocated")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	var machineDeployments []v1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&metav1.LabelSelector{},
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

	{
		// Acquire AZs together with corresponding subnets from AWS EC2 API
		// results.
		azs, err = fromEC2SubnetsToMap(cc.Status.TenantCluster.TCCP.Subnets)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		// Acquire AZs together with corresponding subnets from AWS EC2 API
		// results.
		azs, err = fromEC2RouteTablesToMap(cc.Status.TenantCluster.TCCP.RouteTables)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		status := newAZStatus(azs)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("AZs for cc status: %s", azSubnetsToString(azs)))

		// Add the current AZ state from AWS to the cc status.
		cc.Status.TenantCluster.TCCP.AvailabilityZones = status
	}

	{
		// Parse TCCP network CIDR.
		_, tccpSubnet, err := net.ParseCIDR(key.StatusClusterNetworkCIDR(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		azs, err = r.ensureAZsAreAssignedWithSubnet(ctx, *tccpSubnet, azs)
		if err != nil {
			return microerror.Mask(err)
		}

		spec := newAZSpec(azs)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("AZs for cc spec: %s", azSubnetsToString(azs)))

		// Add the desired AZ state to the controllercontext spec.
		cc.Spec.TenantCluster.TCCP.AvailabilityZones = spec
	}

	return nil
}

// ensureAZsAreAssignedWithSubnet iterates over AZ-subnetPair map, removes
// subnets that are already in use from available subnets' list and then assigns
// one to AZs that doesn't have subnet assigned yet.
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
		for az := range azs {
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
			parentNet := ipam.CalculateParent(subnets.Public.CIDR)

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

				subnets.Public.CIDR = ss[0]
				subnets.Private.CIDR = ss[1]
				azs[az] = subnets
			} else {
				return nil, microerror.Maskf(invalidConfigError, "no more unallocated subnets left but there's this AZ still left: %q", az)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting cluster availability zones to controllercontext: %#v", azs))

	return azs, nil
}

func azSubnetsToString(azs map[string]subnetPair) string {
	var result strings.Builder
	result.WriteString("availability zone subnet allocations: {")
	for az, subnet := range azs {
		result.WriteString(fmt.Sprintf("\n%q: [pub: %q, private: %q]", az, subnet.Public.CIDR.String(), subnet.Private.CIDR.String()))
	}

	if len(azs) > 0 {
		result.WriteString("\n}")
	} else {
		result.WriteString("}")
	}

	return result.String()
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
			mappedSubnet.Public.ID = *s.SubnetId
			mappedSubnet.Public.CIDR = *cidr
		case "private":
			mappedSubnet.Private.ID = *s.SubnetId
			mappedSubnet.Private.CIDR = *cidr
		default:
			return nil, microerror.Maskf(invalidConfigError, "invalid subnet type in ec2.Subnet tag: %q: %q", key.TagSubnetType, subnetType)
		}

		azMap[*s.AvailabilityZone] = mappedSubnet
	}

	return azMap, nil
}

func newAZSpec(azs map[string]subnetPair) []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone {
	var azNames []string
	{
		for az := range azs {
			azNames = append(azNames, az)
		}

		sort.Strings(azNames)
	}

	var spec []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone

	for _, name := range azNames {
		sp := azs[name]

		az := controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
			Name: name,
			Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
				Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
					CIDR: sp.Private.CIDR,
					ID:   sp.Private.ID,
				},
				Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
					CIDR: sp.Public.CIDR,
					ID:   sp.Public.ID,
				},
			},
		}

		spec = append(spec, az)
	}

	return spec
}

func newAZStatus(azs map[string]subnetPair) []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone {
	var azNames []string
	{
		for az := range azs {
			azNames = append(azNames, az)
		}

		sort.Strings(azNames)
	}

	var status []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone

	for _, name := range azNames {
		sp := azs[name]

		// Skip empty subnets as they are not allocated in AWS and therefor not in
		// the current state.
		if sp.areEmpty() {
			continue
		}

		// Collect currently used AZ information to store it inside the cc status.
		az := controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
			Name: name,
			Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
				Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
					CIDR: sp.Private.CIDR,
					ID:   sp.Private.ID,
				},
				Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
					CIDR: sp.Public.CIDR,
					ID:   sp.Public.ID,
				},
			},
		}

		status = append(status, az)
	}

	return status
}
