package clusterazs

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"

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

const MaximumNumberOfAZsInCluster = 4

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

	var azs []string
	{
		var azsMap map[string]struct{}

		// Include master's AZ.
		azsMap[key.MasterAvailabilityZone(cr)] = struct{}{}

		// Include worker AZs.
		for _, md := range machineDeployments {
			for _, az := range key.WorkerAvailabilityZones(md) {
				azsMap[az] = struct{}{}
			}
		}

		for az := range azsMap {
			azs = append(azs, az)
		}
	}

	{
		// Temporary type for mapping existing subnets from controllercontext to AZs.
		type subnet struct {
			public  net.IPNet
			private net.IPNet
		}

		var allocatedNetworks []net.IPNet
		azMap := make(map[string]subnet)

		for _, s := range cc.Status.TenantCluster.TCCP.Subnets {
			if s == nil || s.AvailabilityZone == nil || s.CidrBlock == nil || s.Tags == nil {
				continue
			}

			mappedSubnet := azMap[*s.AvailabilityZone]

			_, cidr, err := net.ParseCIDR(*s.CidrBlock)
			if err != nil {
				return microerror.Mask(err)
			}

			// Maintain list of existing network allocations for cluster
			// network.
			allocatedNetworks = append(allocatedNetworks, *cidr)

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
				return microerror.Maskf(invalidConfigError, "invalid subnet type in ec2.Subnet tag: %q: %q", key.TagEC2SubnetType, subnetType)
			}

			azMap[*s.AvailabilityZone] = mappedSubnet
		}

		_, clusterCIDR, err := net.ParseCIDR(key.StatusClusterNetworkCIDR(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		clusterAZSubnets, err := ipam.Split(*clusterCIDR, MaximumNumberOfAZsInCluster)
		if err != nil {
			return microerror.Mask(err)
		}

		var ccAZs []controllercontext.ContextStatusTenantClusterAvailabilityZone
		for _, az := range azs {
			ccAZ := controllercontext.ContextStatusTenantClusterAvailabilityZone{
				Name: az,
			}

			subnet, exists := azMap[az]
			if exists {
				ccAZ.PublicSubnet = subnet.public
				ccAZ.PrivateSubnet = subnet.private

				parentNet := ipam.CalculateParent(subnet.public)

				clusterAZSubnets = ipam.Filter(clusterAZSubnets, func(n net.IPNet) bool {
					return !reflect.DeepEqual(n, parentNet)
				})
			} else {
				if len(clusterAZSubnets) > 0 {

					subnets, err := ipam.Split(clusterAZSubnets[0], 2)
					if err != nil {
						return microerror.Mask(err)
					}

					clusterAZSubnets = clusterAZSubnets[1:]
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
