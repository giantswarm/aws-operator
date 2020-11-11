package ipam

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/locker"
)

// EnsureCreated allocates tenant cluster network segments. It gathers existing
// subnets from existing system resources like VPCs and Cluster CRs.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "acquiring lock for IPAM")
		err := r.locker.Lock(ctx)
		if locker.IsAlreadyExists(err) {
			// In case the lock already exists we stop here and try again during
			// the next reconciliation loop because another process is already
			// trying to allocate subnets.
			r.logger.LogCtx(ctx, "level", "debug", "message", "lock for IPAM is already acquired")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "acquired lock for IPAM")
		}

		defer func() {
			r.logger.LogCtx(ctx, "level", "debug", "message", "releasing lock for IPAM")
			err := r.locker.Unlock(ctx)
			if locker.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "lock for IPAM is already released")
			} else if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", "failed to release lock for IPAM", "stack", fmt.Sprintf("%#v", err))
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "released lock for IPAM")
			}
		}()
	}

	{
		proceed, err := r.checker.Check(ctx, m.GetNamespace(), m.GetName())
		if err != nil {
			return microerror.Mask(err)
		}

		if !proceed {
			r.logger.LogCtx(ctx, "level", "debug", "message", "subnet already allocated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	// This is the custom network range configured by the NetworkPool CR. Since
	// this is dynamic we need to look it up in order to consider it for network
	// allocation, if the NetworkPool CR is given.
	var networkRange net.IPNet
	{

		var cr v1alpha2.AWSCluster
		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: key.ClusterID(m), Namespace: m.GetNamespace()}, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if cr.Spec.Provider.Nodes.NetworkPool == "" {
			networkRange = r.networkRange
		} else {
			var np v1alpha2.NetworkPool
			err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: cr.Spec.Provider.Nodes.NetworkPool, Namespace: cr.GetNamespace()}, &np)
			if err != nil {
				return microerror.Mask(err)
			}

			_, ipnet, err := net.ParseCIDR(np.Spec.CIDRBlock)
			if err != nil {
				return microerror.Mask(err)
			}
			networkRange = *ipnet
		}
	}

	var allocatedSubnets []net.IPNet
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding allocated subnets")

		allocatedSubnets, err = r.collector.Collect(ctx, networkRange)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found allocated subnets %#q", allocatedSubnets))
	}

	var freeSubnet net.IPNet
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")

		var subnetMask net.IPMask
		{
			subnetMaskString, ok := m.GetAnnotations()[annotation.AWSSubnetSize]
			if ok {
				subnetBits, err := strconv.Atoi(subnetMaskString)
				if err != nil {
					return microerror.Mask(err)
				}
				subnetMask = net.CIDRMask(subnetBits, 32)
			} else {
				subnetMask = r.allocatedSubnetMask
			}
		}

		freeSubnet, err = ipam.Free(networkRange, subnetMask, allocatedSubnets)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet %#q", freeSubnet))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("allocating free subnet %#q", freeSubnet))

		err = r.persister.Persist(ctx, freeSubnet, m.GetNamespace(), m.GetName())
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("allocated free subnet %#q", freeSubnet))
	}

	return nil
}
