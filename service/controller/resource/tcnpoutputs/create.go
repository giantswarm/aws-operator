package tcnpoutputs

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v16/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v16/service/controller/key"
	"github.com/giantswarm/aws-operator/v16/service/internal/cloudformation"
)

const (
	DockerVolumeSizeGBKey = "DockerVolumeSizeGB"
	InstanceImageKey      = "InstanceImage"
	InstanceTypeKey       = "InstanceType"
	OperatorVersionKey    = "OperatorVersion"
	ReleaseVersionKey     = "ReleaseVersion"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
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
		r.logger.Debugf(ctx, "finding the tenant cluster's node pool cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.StackNameTCNP(&cr))
		if cloudformation.IsStackNotFound(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's node pool cloud formation stack outputs")
			r.logger.Debugf(ctx, "the tenant cluster's node pool cloud formation stack does not exist")
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if cloudformation.IsOutputsNotAccessible(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's node pool cloud formation stack outputs")
			r.logger.Debugf(ctx, "the tenant cluster's node pool cloud formation stack output values are not accessible due to stack status %#q", s)
			r.logger.Debugf(ctx, "canceling resource")
			cc.Status.TenantCluster.TCCP.IsTransitioning = true
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.Debugf(ctx, "found the tenant cluster's node pool cloud formation stack outputs")
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, DockerVolumeSizeGBKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, InstanceImageKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCNP.WorkerInstance.Image = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, InstanceTypeKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCNP.WorkerInstance.Type = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, OperatorVersionKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.OperatorVersion = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, ReleaseVersionKey)
		if cloudformation.IsOutputNotFound(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane nodes ReleaseVersion output")
		} else if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.ReleaseVersion = v
	}

	return nil
}
