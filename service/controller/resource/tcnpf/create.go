package tcnpf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnpf/template"
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

	{
		if len(cc.Spec.TenantCluster.TCNP.AvailabilityZones) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "availability zone information not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		if cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the VPC Peering Connection ID in the controller context")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's node pool finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCNPF(&cr)),
		}

		o, err := cc.Client.ControlPlane.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			// fall through

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(eventCFCreateError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusRollbackFailed {
			return microerror.Maskf(eventCFRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateRollbackFailed {
			return microerror.Maskf(eventCFUpdateRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's node pool finalizer cloud formation stack already exists")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's node pool finalizer cloud formation stack")
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's node pool finalizer cloud formation stack")

		params, err := newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's node pool finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's node pool finalizer cloud formation stack")

		i := &cloudformation.CreateStackInput{
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCNPF(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.ControlPlane.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's node pool finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for the creation of the tenant cluster's node pool finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCNPF(&cr)),
		}

		err = cc.Client.ControlPlane.AWS.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "waited for the creation of the tenant cluster's node pool finalizer cloud formation stack")
	}

	return nil
}

func newRouteTablesParams(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainRouteTables, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var peeringConnections []template.ParamsMainVPCPeeringConnection
	for _, rt := range cc.Status.ControlPlane.RouteTables {
		for _, az := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
			pc := template.ParamsMainVPCPeeringConnection{
				ID:   cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
				Name: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name), awstags.ValueForKey(rt.Tags, "Name")),
				RouteTable: template.ParamsMainVPCPeeringConnectionRouteTable{
					ID: *rt.RouteTableId,
				},
				Subnet: template.ParamsMainVPCPeeringConnectionSubnet{
					CIDR: az.Subnet.Private.CIDR.String(),
				},
			}

			peeringConnections = append(peeringConnections, pc)
		}
	}

	var routeTables *template.ParamsMainRouteTables
	{
		routeTables = &template.ParamsMainRouteTables{
			PeeringConnections: peeringConnections,
		}
	}

	return routeTables, nil
}

func newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		routeTables, err := newRouteTablesParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			RouteTables: routeTables,
		}
	}

	return params, nil
}
