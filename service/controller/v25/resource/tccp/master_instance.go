package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) searchMasterInstanceID(ctx context.Context, cr v1alpha1.AWSConfig) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var instanceID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding master instance ID for %#q", key.MasterInstanceName(cr)))

		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.MasterInstanceName(cr)),
					},
				},
				{
					Name: aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{
						aws.String(key.ClusterID(cr)),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
		if err != nil {
			return "", microerror.Mask(err)
		}

		if len(o.Reservations) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations))
		}
		if len(o.Reservations[0].Instances) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations[0].Instances))
		}

		instanceID = *o.Reservations[0].Instances[0].InstanceId

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found master instance ID %#q for %#q", instanceID, key.MasterInstanceName(cr)))
	}

	return instanceID, nil
}

func (r *Resource) terminateMasterInstance(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	instanceID, err := r.searchMasterInstanceID(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("disabling termination protection for master instance %#q", instanceID))

		i := &ec2.ModifyInstanceAttributeInput{
			DisableApiTermination: &ec2.AttributeBooleanValue{
				Value: aws.Bool(false),
			},
			InstanceId: aws.String(instanceID),
		}

		_, err = cc.Client.TenantCluster.AWS.EC2.ModifyInstanceAttribute(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("disabled termination protection for master instance %#q", instanceID))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminating master instance %#q", instanceID))

		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.TerminateInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminated master instance %#q", instanceID))
	}

	return nil
}
