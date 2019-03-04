package stackoutput

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	cf "github.com/giantswarm/aws-operator/service/controller/v24/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var outputs []*cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster cloud formation stack outputs")

		o, s, err := cc.CloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(cr))
		if cf.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster cloud formation stack does not exist")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if cf.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster main cloud formation stack output values are not accessible due to stack status %#q", s))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster cloud formation stack outputs")
	}

	if r.route53Enabled {
		v, err := cc.CloudFormation.GetOutputValue(outputs, key.HostedZoneNameServers)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.Cluster.HostedZoneNameServers = v
	}

	{
		v, err := cc.CloudFormation.GetOutputValue(outputs, key.VPCPeeringConnectionIDKey)
		if cf.IsStackNotFound(err) {
			// TODO this exception is necessary for clusters upgrading from v23 to
			// v24. The code can be cleaned up in v25 and the controller context value
			// assignment can be managed like the other examples below.
			//
			//     https://github.com/giantswarm/giantswarm/issues/5496
			//
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			cc.Status.Cluster.VPCPeeringConnectionID = v
		}
	}

	{
		v, err := cc.CloudFormation.GetOutputValue(outputs, key.WorkerASGKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.Drainer.WorkerASGName = v
	}

	return nil
}
