package tccpnoutputs

import (
	"context"
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudformation"
)

const (
	InstanceTypeKey    = "InstanceType"
	OperatorVersionKey = "OperatorVersion"
	MasterReplicasKey  = "MasterReplicas"
	ReleaseVersionKey  = "ReleaseVersion"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToControlPlane(obj)
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
		r.logger.Debugf(ctx, "finding the tenant cluster's control plane nodes cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.StackNameTCCPN(&cr))
		if cloudformation.IsStackNotFound(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane nodes cloud formation stack outputs")
			r.logger.Debugf(ctx, "the tenant cluster's control plane nodes cloud formation stack does not exist")
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if cloudformation.IsOutputsNotAccessible(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane nodes cloud formation stack outputs")
			r.logger.Debugf(ctx, "the tenant cluster's control plane nodes cloud formation stack output values are not accessible due to stack status %#q", s)
			r.logger.Debugf(ctx, "canceling resource")
			cc.Status.TenantCluster.TCCPN.IsTransitioning = true
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.Debugf(ctx, "found the tenant cluster's control plane nodes cloud formation stack outputs")
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, InstanceTypeKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCCPN.InstanceType = v
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

	{
		v, err := cloudFormation.GetOutputValue(outputs, MasterReplicasKey)
		if cloudformation.IsOutputNotFound(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane nodes MasterReplicas output")
		} else if err != nil {
			return microerror.Mask(err)
		}

		// TODO this is migration code and can be removed when the aws-operator got
		// graduated to v9.0.0, because then we can be sure all Tenant Clusters have
		// the proper HA Masters setup including the new stack outputs.
		//
		//     https://github.com/giantswarm/giantswarm/issues/10139
		//
		if v == "" {
			v = "0"
		}

		i, err := strconv.Atoi(v)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.TCCPN.MasterReplicas = i
	}

	return nil
}
