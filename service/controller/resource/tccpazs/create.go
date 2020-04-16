package tccpazs

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sort"

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
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not available yet")
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

		for _, az := range status {
			r.logger.LogCtx(ctx,
				"level", "debug",
				"message", "computed controller context status",
				"availability-zone", az.Name,
				"subnet-id", az.Subnet.Public.CIDR,
				"public-subnet", az.Subnet.Public.CIDR,
				"private-subnet", az.Subnet.Private.CIDR,
				"aws-cni-subnet", az.Subnet.AWSCNI.CIDR,
			)
		}

		// Add the current AZ state from AWS to the cc status.
		cc.Status.TenantCluster.TCCP.AvailabilityZones = status
	}

	{
		// Allow the actual VPC subnet CIDR to be overwritten by the CR spec.
		podSubnet := r.cidrBlockAWSCNI
		if cr.Spec.Provider.Pods.CIDRBlock != "" {
			podSubnet = cr.Spec.Provider.Pods.CIDRBlock
		}

		_, awsCNISubnet, err := net.ParseCIDR(podSubnet)
		if err != nil {
			return microerror.Mask(err)
		}

		// Parse TCCP network CIDR.
		_, tccpSubnet, err := net.ParseCIDR(key.StatusClusterNetworkCIDR(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		azMapping, err = r.ensureAZsAreAssignedWithSubnet(ctx, *awsCNISubnet, *tccpSubnet, azMapping)
		if err != nil {
			return microerror.Mask(err)
		}

		spec := newAZSpec(azMapping)

		for _, az := range spec {
			r.logger.LogCtx(ctx,
				"level", "debug",
				"message", "computed controller context spec",
				"availability-zone", az.Name,
				"subnet-id", az.Subnet.Public.CIDR,
				"public-subnet", az.Subnet.Public.CIDR,
				"private-subnet", az.Subnet.Private.CIDR,
				"aws-cni-subnet", az.Subnet.AWSCNI.CIDR,
			)
		}

		// Add the desired AZ state to the controllercontext spec.
		cc.Spec.TenantCluster.TCCP.AvailabilityZones = spec
	}

	return nil
}

// ensureAZsAreAssignedWithSubnet iterates over AZ-mapping map, removes
// subnets that are already in use from available subnets' list and then assigns
// one to AZs that doesn't have subnet assigned yet.
func (r *Resource) ensureAZsAreAssignedWithSubnet(ctx context.Context, awsCNISubnet net.IPNet, tccpSubnet net.IPNet, azMapping map[string]mapping) (map[string]mapping, error) {
	// Split TCCP network between maximum number of AZs. This is because of
	// current limitation in IPAM design and AWS TCCP infrastructure
	// design.
	clusterAZSubnets, err := ipam.Split(tccpSubnet, MaxAZs)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// According to the IPAM limitations we split the AWS CNI CIDR by 4. This is
	// so we assign one of its split to each availability zone.
	awsCNISubnets, err := ipam.Split(awsCNISubnet, MaxAZs)
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
	// remaining subnets to AZs without them.
	for _, az := range azNames {
		mapping := azMapping[az]

		// Check if mapping of given availability zone already contain value.
		if !mapping.PublicSubnetEmpty() {
			// Calculate the parent network from public subnet (always present for
			// functional AZ).
			parent := ipam.CalculateParent(mapping.Public.Subnet.CIDR)

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %#q has public and private subnet %#q already allocated", az, parent.String()))

			// Filter out already allocated AZ subnet from available AZ size networks.
			clusterAZSubnets = ipam.Filter(clusterAZSubnets, func(n net.IPNet) bool {
				return !reflect.DeepEqual(n, parent)
			})
		}

		// Unlike the public/private subnets the AWS CNI CIDRs are not split twice.
		// We only have at max 4 and not 8. This means we do not have to filter out
		// the parents, but the subnets being taken already themselves.
		if !mapping.AWSCNISubnetEmpty() {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %#q has aws-cni subnet %#q already allocated", az, mapping.AWSCNI.Subnet.CIDR.String()))

			// Filter out already allocated AZ subnet from available AZ size networks.
			awsCNISubnets = ipam.Filter(awsCNISubnets, func(n net.IPNet) bool {
				return !reflect.DeepEqual(n, mapping.AWSCNI.Subnet.CIDR)
			})
		}
	}

	// Assign subnet to AZs that don't it yet.
	for _, az := range azNames {
		mapping := azMapping[az]

		if mapping.PublicSubnetEmpty() && mapping.PrivateSubnetEmpty() {
			if len(clusterAZSubnets) > 0 {
				// Pick first available AZ subnet and split it to public and private.
				clusterAZSubnet, err := ipam.Split(clusterAZSubnets[0], 2)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q doesn't have public and private subnet allocation", az))

				// Update available AZ mapping by removing the CIDR we are about to
				// allocate.
				clusterAZSubnets = clusterAZSubnets[1:]

				mapping.Public.Subnet.CIDR = clusterAZSubnet[0]
				mapping.Private.Subnet.CIDR = clusterAZSubnet[1]

				azMapping[az] = mapping
			} else {
				return nil, microerror.Maskf(invalidConfigError, "no more unallocated subnets left but there's this AZ still left: %q", az)
			}
		}

		if mapping.AWSCNISubnetEmpty() {
			if len(awsCNISubnets) > 0 {
				// The AWS CNI CIDRs are not divided into public/private per AZ. We only
				// split by the IPAM limit. So here we simply take the first free CIDR
				// and remove it below.
				awsCNISubnet := awsCNISubnets[0]

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("availability zone %q doesn't have aws-cni subnet allocation", az))

				// Update available AZ mapping by removing the CIDR we are about to
				// allocate.
				awsCNISubnets = awsCNISubnets[1:]

				mapping.AWSCNI.Subnet.CIDR = awsCNISubnet

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
		case t == "aws-cni":
			m.AWSCNI.Subnet.ID = *s.SubnetId
			m.AWSCNI.Subnet.CIDR = *cidr
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
				AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
					CIDR: sp.AWSCNI.Subnet.CIDR,
					ID:   sp.AWSCNI.Subnet.ID,
				},
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
		if sp.PublicSubnetEmpty() && sp.PrivateSubnetEmpty() && sp.AWSCNISubnetEmpty() {
			continue
		}

		// Collect currently used AZ information to store it inside the cc status.
		az := controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
			Name: name,
			Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
				AWSCNI: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
					CIDR: sp.AWSCNI.Subnet.CIDR,
					ID:   sp.AWSCNI.Subnet.ID,
				},
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
