package cleanuploadbalancers

import (
	"context"
	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

// EnsureDeleted ensures that any ELBs from Kubernetes LoadBalancer services
// are deleted. This is needed because the use the VPC public subnet.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// delete classic load-balancers
	{
		lbState, err := r.clusterClassicLoadBalancers(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if lbState != nil && len(lbState.LoadBalancerNames) > 0 {
			r.logger.Debugf(ctx, "deleting %d load balancers", len(lbState.LoadBalancerNames))

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

			r.logger.Debugf(ctx, "deleted %d load balancers", len(lbState.LoadBalancerNames))
		} else {
			r.logger.Debugf(ctx, "not deleting load balancers because there aren't any")
		}
	}

	// delete load-balancers V2
	{
		lbState, err := r.clusterLoadBalancersV2(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if lbState != nil && len(lbState.LoadBalancerArns) > 0 {
			r.logger.Debugf(ctx, "deleting %d load balancers", len(lbState.LoadBalancerArns))

			cc, err := controllercontext.FromContext(ctx)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, lbArn := range lbState.LoadBalancerArns {
				_, err := cc.Client.TenantCluster.AWS.ELBv2.DeleteLoadBalancer(&elbv2.DeleteLoadBalancerInput{
					LoadBalancerArn: aws.String(lbArn),
				})
				if err != nil {
					return microerror.Mask(err)
				}
			}

			r.logger.Debugf(ctx, "deleted %d load balancers", len(lbState.LoadBalancerArns))
		} else {
			r.logger.Debugf(ctx, "not deleting load balancers because there aren't any")
		}
	}

	return nil
}
