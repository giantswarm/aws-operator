package tccpdetachlbsubnet

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/controllercontext"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpdetachlbsubnetv31"
)

type Config struct {
	Logger micrologger.Logger
}

type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureUnusedAZsAreDetachedFromLBs(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// The tenant cluster VPC is a requirement to find its associated load
	// balancers. In case the VPC ID is not yet tracked in the controller
	// context we return an error and cause the resource to be canceled.
	if cc.Status.TenantCluster.TCCP.VPC.ID == "" {
		return microerror.Mask(vpcNotFoundError)
	}

	// Collect all LBs from current TCCP VPC.
	var elbs []*elb.LoadBalancerDescription
	{
		i := &elb.DescribeLoadBalancersInput{}

		var o *elb.DescribeLoadBalancersOutput

		for o == nil || (o.NextMarker != nil && *o.NextMarker != "") {
			o, err = cc.Client.TenantCluster.AWS.ELB.DescribeLoadBalancers(i)
			if err != nil {
				return microerror.Mask(err)
			}

			if o == nil {
				return microerror.Maskf(executionFailedError, "DescribeLoadBalancersOutput is nil")
			}

			// Copy marker in case there's next page.
			i.Marker = o.NextMarker

			for _, lb := range o.LoadBalancerDescriptions {
				if lb == nil {
					continue
				}

				if lb.VPCId == nil || *lb.VPCId != cc.Status.TenantCluster.TCCP.VPC.ID {
					// This LB doesn't belong to our VPC.
					continue
				}

				elbs = append(elbs, lb)
			}
		}
	}

	// Collect all existing subnets that should be detached.
	var subnetsToDetach []*ec2.Subnet
	{
		for _, snet := range cc.Status.TenantCluster.TCCP.Subnets {
			if snet == nil {
				continue
			}

			var found bool
			for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
				if snet.AvailabilityZone != nil && *snet.AvailabilityZone == az.Name {
					found = true
					break
				}
			}

			// Subnet's AZ is not in TCCP AZs Spec. This means that this AZ is
			// not needed anymore and therefore this subnet must be detached
			// from LBs as well.
			if !found {
				subnetsToDetach = append(subnetsToDetach, snet)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if there are subnets to detach from load balancers")

	// Collect all LB detachments.
	var lbDetachments []*elb.DetachLoadBalancerFromSubnetsInput
	{
		for _, lb := range elbs {
			lbDetachment := &elb.DetachLoadBalancerFromSubnetsInput{
				LoadBalancerName: lb.LoadBalancerName,
			}

			// Iterate over LB subnets in case it contains a subnet that must
			// be detached. If that is the case, we add it to LB detachments.
			for _, snetID := range lb.Subnets {
				for _, snet := range subnetsToDetach {
					if *snet.SubnetId == *snetID {
						lbDetachment.Subnets = append(lbDetachment.Subnets, snet.SubnetId)
					}
				}
			}

			if len(lbDetachment.Subnets) > 0 {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found subnets %#q to be detached from load balancer %#q", aws.StringValueSlice(lbDetachment.Subnets), aws.StringValue(lbDetachment.LoadBalancerName)))

				lbDetachments = append(lbDetachments, lbDetachment)
			}
		}
	}

	if len(lbDetachments) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find subnets to detach from load balancers")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	// Perform the actual detachment.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "detaching subnets from load balancers")

		for _, i := range lbDetachments {
			_, err := cc.Client.TenantCluster.AWS.ELB.DetachLoadBalancerFromSubnets(i)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detached subnets %#q from load balancer %#q", aws.StringValueSlice(i.Subnets), aws.StringValue(i.LoadBalancerName)))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "finished detaching from load balancers")
	}

	return nil
}
