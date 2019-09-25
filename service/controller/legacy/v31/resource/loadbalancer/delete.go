package loadbalancer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v31/key"
)

// EnsureDeleted ensures that any ELBs from Kubernetes LoadBalancer services
// are deleted. This is needed because the use the VPC public subnet.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	lbState, err := r.clusterLoadBalancers(ctx, customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	if lbState != nil && len(lbState.LoadBalancerNames) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %d load balancers", len(lbState.LoadBalancerNames)))

		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, lbName := range lbState.LoadBalancerNames {
			_, err := cc.Client.TenantCluster.AWS.ELB.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{
				LoadBalancerName: aws.String(lbName),
			})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted %d load balancers", len(lbState.LoadBalancerNames)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting load balancers because there aren't any")
	}

	return nil
}
