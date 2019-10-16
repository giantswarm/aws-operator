package tccpingresstargetgroups

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

const (
	IngressInsecureTargetGroupIDs = "IngressInsecureTargetGroupsID"
	IngressSecureTargetGroupIDs   = "IngressSecureTargetGroupsID"
)

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

	var cloudFormation *cloudformation.CloudFormation
	{
		c := cloudformation.Config{
			Client: cc.Client.TenantCluster.AWS.CloudFormation,
		}

		cloudFormation, err = cloudformation.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var outputs []cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.StackNameTCCP(&cr))
		if cloudformation.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane cloud formation stack does not exist")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if cloudformation.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack output values are not accessible due to stack status %#q", s))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			cc.Status.TenantCluster.TCCP.IsTransitioning = true
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane cloud formation stack outputs")
	}

	{
		ingressInsecureTargetGroup, err := cloudFormation.GetOutputValue(outputs, IngressInsecureTargetGroupIDs)
		if err != nil {
			return microerror.Mask(err)
		}
		ingressSecureTargetGroup, err := cloudFormation.GetOutputValue(outputs, IngressSecureTargetGroupIDs)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCCP.IngressTargetGroupIDs = []string{ingressInsecureTargetGroup, ingressSecureTargetGroup}
	}

	return nil
}
