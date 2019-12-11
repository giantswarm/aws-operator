package tccpazs

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

// MaxAZs is the maximum number of availability zones allowed for a tenant
// cluster. The major factor causing this limitation is the current IPAM
// implementation. It restricts network sizes in a certain way. Another related
// problem is restrictions in AWS resource structure.
const MaxAZs = 4

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not yet availabile")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
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

	var machineDeployments []infrastructurev1alpha2.AWSMachineDeployment
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

		list, err := r.g8sClient.InfrastructureV1alpha2().AWSMachineDeployments(metav1.NamespaceAll).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		machineDeployments = list.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments)))
	}

	// As a first step we initialize the mappings for all the relevant
	// availability zone information coming from the master node, the worker nodes
	// and the EC2 subnets. Having all the availability zones inside the map makes
	// it easier to fill their corresponding subnet and route table information
	// below.
	azMapping := map[string]mapping{}
	{
		azMapping[key.MasterAvailabilityZone(cr)] = mapping{
			RequiredByCR: true,
		}

		for _, md := range machineDeployments {
			for _, az := range key.MachineDeploymentAvailabilityZones(md) {
				azMapping[az] = mapping{
					RequiredByCR: true,
				}
			}
		}

		for _, az := range azsFromSubnets(cc.Status.TenantCluster.TCCP.Subnets) {
			_, exists := azMapping[az]
			if !exists {
				azMapping[az] = mapping{
					RequiredByCR: false,
				}
			}
		}
	}

	// Map subnet information to their corresponding availability zones based on
	// the AWS API results.
	{
		azMapping, err = mapSubnets(azMapping, cc.Status.TenantCluster.TCCP.Subnets)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		status := newAZStatus(azMapping)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("AZs for cc status: %s", azSubnetStatusToString(status)))

		// Add the current AZ state from AWS to the cc status.
		cc.Status.TenantCluster.TCCP.AvailabilityZones = status
	}

	{
		// Parse TCCP network CIDR.
		_, tccpSubnet, err := net.ParseCIDR(key.StatusClusterNetworkCIDR(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		azMapping, err = r.ensureAZsAreAssignedWithSubnet(ctx, *tccpSubnet, azMapping)
		if err != nil {
			return microerror.Mask(err)
		}

		spec := newAZSpec(azMapping)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("AZs for cc spec: %s", azSubnetSpecToString(spec)))

		// Add the desired AZ state to the controllercontext spec.
		cc.Spec.TenantCluster.TCCP.AvailabilityZones = spec
	}

	return nil
}

// ensureAZsAreAssignedWithSubnet iterates over AZ-mapping map, removes
// subnets that are already in use from available subnets' list and then assigns
// one to AZs that doesn't have subnet assigned yet.
func (r *Resource) ensureAZsAreAssignedWithSubnet(ctx context.Context, tccpSubnet net.IPNet, azMapping map[string]mapping) (map[string]mapping, error) {
	// Split TCCP network between maximum number of AZs. This is because of
	// current limitation in IPAM design and AWS TCCP infrastructure
	// design.
	clusterAZSubnets, err := ipam.Split(tccpSubnet, MaxAZs)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var azNames []string
	{
		for az := range azMapping {
			azNames = append(azNames, az)
		}

		sort.Strings(azNames)
	}

	// Remove already allocated networks from clusterAZSubnets before assigning
	// remaining subnets to AZs without themout them.
	for _, az := range azNames {
		mapping := azMapping[az]

		// Check if mapping of given availability zone already contain value.
		if !mapping.subnetsEmpty() {
			// Calculate the parent network from public subnet (always
			// present for functional AZ).
			parentNet := ipam.CalculateParent(mapping.Public.Subnet.CIDR)

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
		mapping := azMapping[az]

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
				azMapping[az] = mapping
			} else {
				return nil, microerror.Maskf(invalidConfigError, "no more unallocated subnets left but there's this AZ still left: %q", az)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("AZ subnet mappings: %#v", azMapping))

	return azMapping, nil
}

func azsFromSubnets(subnets []*ec2.Subnet) []string {
	var azs []string

	for _, s := range subnets {
		azs = append(azs, *s.AvailabilityZone)
	}

	return azs
}

func azSubnetSpecToString(azs []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone) string {
	var result strings.Builder
	result.WriteString("availability zone subnet allocations: {")
	for _, az := range azs {
		result.WriteString(fmt.Sprintf("\n%q: [pub: %q, private: %q]", az.Name, az.Subnet.Public.CIDR.String(), az.Subnet.Private.CIDR.String()))
	}

	if len(azs) > 0 {
		result.WriteString("\n}")
	} else {
		result.WriteString("}")
	}

	return result.String()
}

func azSubnetStatusToString(azs []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) string {
	var result strings.Builder
	result.WriteString("availability zone subnet allocations: {")
	for _, az := range azs {
		result.WriteString(fmt.Sprintf("\n%q: [pub: %q, private: %q]", az.Name, az.Subnet.Public.CIDR.String(), az.Subnet.Private.CIDR.String()))
	}

	if len(azs) > 0 {
		result.WriteString("\n}")
	} else {
		result.WriteString("}")
	}

	return result.String()
}

func mapSubnets(azMapping map[string]mapping, subnets []*ec2.Subnet) (map[string]mapping, error) {
	for _, s := range subnets {
		if !awstags.HasTags(s.Tags, key.TagSubnetType) {
			// Filter out EC2 subnets that don't specify subnet type (i.e.
			// public or private).
			continue
		}
		if awstags.ValueForKey(s.Tags, key.TagStack) != key.StackTCCP {
			// Filter out EC2 subnets that don't belong to tenant cluster
			// control plane CF stack.
			continue
		}

		_, cidr, err := net.ParseCIDR(*s.CidrBlock)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		m := azMapping[*s.AvailabilityZone]

		switch t := awstags.ValueForKey(s.Tags, key.TagSubnetType); {
		case t == "public":
			m.Public.Subnet.ID = *s.SubnetId
			m.Public.Subnet.CIDR = *cidr
		case t == "private":
			m.Private.Subnet.ID = *s.SubnetId
			m.Private.Subnet.CIDR = *cidr
		default:
			return nil, microerror.Maskf(invalidConfigError, "invalid subnet type %#q", t)
		}

		azMapping[*s.AvailabilityZone] = m
	}

	return azMapping, nil
}

func newAZSpec(azMapping map[string]mapping) []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone {
	var azNames []string
	{
		for az := range azMapping {
			azNames = append(azNames, az)
		}

		sort.Strings(azNames)
	}

	var spec []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone

	for _, name := range azNames {
		sp := azMapping[name]

		az := controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
			Name: name,
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

func newAZStatus(azMapping map[string]mapping) []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone {
	var azNames []string
	{
		for az := range azMapping {
			azNames = append(azNames, az)
		}

		sort.Strings(azNames)
	}

	var status []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone

	for _, name := range azNames {
		sp := azMapping[name]

		// Skip empty subnets as they are not allocated in AWS and therefor not in
		// the current state.
		if sp.subnetsEmpty() {
			continue
		}

		// Collect currently used AZ information to store it inside the cc status.
		az := controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
			Name: name,
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
