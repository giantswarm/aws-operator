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

	azs := map[string]mapping{}
	{
		// Include master's AZ if missing.
		if _, exists := azs[key.MasterAvailabilityZone(cr)]; !exists {
			azs[key.MasterAvailabilityZone(cr)] = mapping{}
		}

		// Include worker AZs if missing.
		for _, md := range machineDeployments {
			for _, az := range key.WorkerAvailabilityZones(md) {
				if _, exists := azs[az]; !exists {
					azs[az] = mapping{}
				}
			}
		}
	}

	// Map subnet and route table information to their corresponding availability
	// zones based on the AWS API results.
	{
		azs, err = mapSubnets(azs, cc.Status.TenantCluster.TCCP.Subnets)
		if err != nil {
			return microerror.Mask(err)
		}

		azs, err = mapRouteTables(azs, cc.Status.TenantCluster.TCCP.RouteTables)
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

// ensureAZsAreAssignedWithSubnet iterates over AZ-mapping map, removes
// subnets that are already in use from available subnets' list and then assigns
// one to AZs that doesn't have subnet assigned yet.
func (r *Resource) ensureAZsAreAssignedWithSubnet(ctx context.Context, tccpSubnet net.IPNet, azs map[string]mapping) (map[string]mapping, error) {
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
		if !subnets.subnetsEmpty() {
			// Calculate the parent network from public subnet (always
			// present for functional AZ).
			parentNet := ipam.CalculateParent(subnets.Public.Subnet.CIDR)

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
		mapping := azs[az]

		// Only proceed with AZs that don't have subnet allocated yet.
		if mapping.subnetsEmpty() {
			if len(clusterAZSubnets) > 0 {
				// Pick first available AZ subnet and split it to public
				// and private.
				ss, err := ipam.Split(clusterAZSubnets[0], 2)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q doesn't have subnet allocation - allocated: %q", az, clusterAZSubnets[0].String()))

				// Update available AZ mapping.
				clusterAZSubnets = clusterAZSubnets[1:]

				mapping.Public.Subnet.CIDR = ss[0]
				mapping.Private.Subnet.CIDR = ss[1]
				azs[az] = mapping
			} else {
				return nil, microerror.Maskf(invalidConfigError, "no more unallocated subnets left but there's this AZ still left: %q", az)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting cluster availability zones to controllercontext: %#v", azs))

	return azs, nil
}

func azSubnetsToString(azs map[string]mapping) string {
	var result strings.Builder
	result.WriteString("availability zone subnet allocations: {")
	for az, mapping := range azs {
		result.WriteString(fmt.Sprintf("\n%q: [pub: %q, private: %q]", az, mapping.Public.Subnet.CIDR.String(), mapping.Private.Subnet.CIDR.String()))
	}

	if len(azs) > 0 {
		result.WriteString("\n}")
	} else {
		result.WriteString("}")
	}

	return result.String()
}

func hasTag(tags []*ec2.Tag, key string) bool {
	for _, t := range tags {
		if *t.Key == key {
			return true
		}
	}

	return false
}

func hasTags(tags []*ec2.Tag, keys ...string) bool {
	for _, k := range keys {
		if !hasTag(tags, k) {
			return false
		}
	}

	return true
}

func mapRouteTables(azs map[string]mapping, routeTables []*ec2.RouteTable) (map[string]mapping, error) {
	for _, rt := range routeTables {
		if !hasTags(rt.Tags, key.TagTCCP, key.TagRouteTableType) {
			continue
		}

		for az, m := range azs {
			if valueForKey(rt.Tags, key.TagAvailabilityZone) != az {
				continue
			}

			switch t := valueForKey(rt.Tags, key.TagRouteTableType); {
			case t == "public":
				m.Public.RouteTable.ID = *rt.RouteTableId
			case t == "private":
				m.Private.RouteTable.ID = *rt.RouteTableId
			default:
				return nil, microerror.Maskf(invalidConfigError, "invalid route table type %#q", t)
			}
		}
	}

	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("xh3b4sd\n")
	for az, m := range azs {
		fmt.Printf("%#v %#v\n", az, m)
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")

	return nil, nil
}

func mapSubnets(azs map[string]mapping, subnets []*ec2.Subnet) (map[string]mapping, error) {
	for _, s := range subnets {
		if !hasTags(s.Tags, key.TagTCCP, key.TagSubnetType) {
			continue
		}

		_, cidr, err := net.ParseCIDR(*s.CidrBlock)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		m := azs[*s.AvailabilityZone]

		switch t := valueForKey(s.Tags, key.TagSubnetType); {
		case t == "public":
			m.Public.Subnet.ID = *s.SubnetId
			m.Public.Subnet.CIDR = *cidr
		case t == "private":
			m.Private.Subnet.ID = *s.SubnetId
			m.Private.Subnet.CIDR = *cidr
		default:
			return nil, microerror.Maskf(invalidConfigError, "invalid subnet type %#q", t)
		}

		azs[*s.AvailabilityZone] = m
	}

	return azs, nil
}

func newAZSpec(azs map[string]mapping) []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone {
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
			RouteTable: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTable{
				Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePrivate{
					ID: sp.Private.RouteTable.ID,
				},
				Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic{
					ID: sp.Public.RouteTable.ID,
				},
			},
			Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
				Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
					CIDR: sp.Private.Subnet.CIDR,
					ID:   sp.Private.Subnet.ID,
				},
				Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
					CIDR: sp.Public.Subnet.CIDR,
					ID:   sp.Public.Subnet.ID,
				},
			},
		}

		spec = append(spec, az)
	}

	return spec
}

func newAZStatus(azs map[string]mapping) []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone {
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
		if sp.subnetsEmpty() {
			continue
		}

		// Collect currently used AZ information to store it inside the cc status.
		az := controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
			Name: name,
			RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
				Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
					ID: sp.Public.RouteTable.ID,
				},
			},
			Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
				Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
					CIDR: sp.Private.Subnet.CIDR,
					ID:   sp.Private.Subnet.ID,
				},
				Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
					CIDR: sp.Public.Subnet.CIDR,
					ID:   sp.Public.Subnet.ID,
				},
			},
		}

		status = append(status, az)
	}

	return status
}

func valueForKey(tags []*ec2.Tag, key string) string {
	for _, t := range tags {
		if *t.Key == key {
			return *t.Value
		}
	}

	return ""
}
