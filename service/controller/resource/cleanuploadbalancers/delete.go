package cleanuploadbalancers

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v2/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v2/service/controller/key"
)

// EnsureDeleted ensures that any ELBs from Kubernetes LoadBalancer services
// are deleted. This is needed because the use the VPC public subnet.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
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

	// delete load-balancers V2 and their target groups
	{
		lbState, err := r.clusterLoadBalancersV2(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		// delete lb v2
		if lbState != nil && len(lbState.LoadBalancerArns) > 0 {
			r.logger.Debugf(ctx, "deleting %d load balancers v2", len(lbState.LoadBalancerArns))

			for _, lbArn := range lbState.LoadBalancerArns {
				_, err := cc.Client.TenantCluster.AWS.ELBv2.DeleteLoadBalancer(&elbv2.DeleteLoadBalancerInput{
					LoadBalancerArn: aws.String(lbArn),
				})
				if err != nil {
					return microerror.Mask(err)
				}
			}

			r.logger.Debugf(ctx, "deleted %d load balancers v2", len(lbState.LoadBalancerArns))
		} else {
			r.logger.Debugf(ctx, "not deleting load balancers v2 because there aren't any")
		}

		// delete target groups
		if lbState != nil && len(lbState.TargetGroupsArns) > 0 {
			r.logger.Debugf(ctx, "deleting %d target groups for load balancers v2", len(lbState.TargetGroupsArns))

			for _, targetGroupArn := range lbState.TargetGroupsArns {
				_, err := cc.Client.TenantCluster.AWS.ELBv2.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
					TargetGroupArn: aws.String(targetGroupArn),
				})
				if err != nil {
					return microerror.Mask(err)
				}
			}

			r.logger.Debugf(ctx, "deleted %d target groups for load balancers v2", len(lbState.TargetGroupsArns))
		} else {
			r.logger.Debugf(ctx, "not deleting target groups load balancers v2 because there aren't any")
		}

	}

	return nil
}
