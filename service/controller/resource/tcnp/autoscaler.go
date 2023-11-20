package tcnp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

func (r *Resource) getASGName(ctx context.Context, cr infrastructurev1alpha3.AWSMachineDeployment) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	find := autoscaling.DescribeAutoScalingGroupsInput{
		Filters: []*autoscaling.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagInstallation)),
				Values: []*string{
					aws.String(r.installationName),
				},
			},
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
				Values: []*string{
					aws.String(key.ClusterID(&cr)),
				},
			},
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagMachineDeployment)),
				Values: []*string{
					aws.String(key.MachineDeploymentID(&cr)),
				},
			},
		},
		MaxRecords: aws.Int64(2),
	}

	// get ASG name
	asgs, err := cc.Client.TenantCluster.AWS.AutoScaling.DescribeAutoScalingGroups(&find)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if len(asgs.AutoScalingGroups) != 1 {
		return "", microerror.Maskf(asgLookupError, "Expected to find exactly 1 ASG, got %d", len(asgs.AutoScalingGroups))
	}

	return *asgs.AutoScalingGroups[0].AutoScalingGroupName, nil
}

func (r *Resource) ensureAutoscalerTag(ctx context.Context, cr infrastructurev1alpha3.AWSMachineDeployment) error {
	r.logger.Debugf(ctx, "Ensuring ASG for nodepool %s has %q tag", cr.Name, clusterAutoscalerEnabledTagName)

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	asgName, err := r.getASGName(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	input := autoscaling.CreateOrUpdateTagsInput{
		Tags: []*autoscaling.Tag{
			{
				Key:               aws.String(clusterAutoscalerEnabledTagName),
				PropagateAtLaunch: aws.Bool(false),
				ResourceId:        aws.String(asgName),
				ResourceType:      aws.String("auto-scaling-group"),
				Value:             aws.String("true"),
			},
		},
	}

	_, err = cc.Client.TenantCluster.AWS.AutoScaling.CreateOrUpdateTags(&input)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "Ensured ASG for nodepool %s has %q tag", cr.Name, clusterAutoscalerEnabledTagName)

	return nil
}

func (r *Resource) removeAutoscalerTag(ctx context.Context, cr infrastructurev1alpha3.AWSMachineDeployment) error {
	r.logger.Debugf(ctx, "Removing tag %q from ASG for nodepool %s", clusterAutoscalerEnabledTagName, cr.Name)

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	asgName, err := r.getASGName(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	input := autoscaling.DeleteTagsInput{
		Tags: []*autoscaling.Tag{
			{
				Key:          aws.String(clusterAutoscalerEnabledTagName),
				ResourceId:   aws.String(asgName),
				ResourceType: aws.String("auto-scaling-group"),
			},
		},
	}

	_, err = cc.Client.TenantCluster.AWS.AutoScaling.DeleteTags(&input)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "Removed tag %q from ASG for nodepool %s", clusterAutoscalerEnabledTagName, cr.Name)

	return nil
}
