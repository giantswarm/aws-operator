package tccpf

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpf/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		if cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the VPC Peering Connection ID in the controller context")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPF(&cr)),
		}

		o, err := cc.Client.ControlPlane.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane finalizer cloud formation stack")
			err = r.createStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane finalizer cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane finalizer cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane finalizer cloud formation stack already exists")
	}

	{
		update, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if update {
			err = r.updateStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (r *Resource) createStack(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane finalizer cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCCPF(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.ControlPlane.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for the creation of the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPF(&cr)),
		}

		err = cc.Client.ControlPlane.AWS.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "waited for the creation of the tenant cluster's control plane finalizer cloud formation stack")
	}

	return nil
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSCluster) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCCPF
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) newRecordSetsParams(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainRecordSets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var recordSets *template.ParamsMainRecordSets
	{
		recordSets = &template.ParamsMainRecordSets{
			BaseDomain:                 key.ClusterBaseDomain(cr),
			ClusterID:                  key.ClusterID(&cr),
			GuestHostedZoneNameServers: cc.Status.TenantCluster.DNS.HostedZoneNameServers,
			Route53Enabled:             r.route53Enabled,
		}
	}

	return recordSets, nil
}

func (r *Resource) newRouteTablesParams(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainRouteTables, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var privateRoutes []template.ParamsMainRouteTablesRoute
	{
		for _, rt := range cc.Status.ControlPlane.RouteTables {
			for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
				// Only those AZs have private subnet in TCCP that run master
				// node. Rest of the AZs are there with public subnet only
				// while the private subnet exists in corresponding node pools.
				// Therefore we need to skip nil Private.CIDRs because there's
				// nothing where we can route the traffic to.
				if az.Subnet.Private.CIDR.IP == nil || az.Subnet.Private.CIDR.Mask == nil {
					continue
				}

				route := template.ParamsMainRouteTablesRoute{
					RouteTableID: *rt.RouteTableId,
					// Requester CIDR block, we create the peering connection from the
					// tenant's private subnets.
					CidrBlock: az.Subnet.Private.CIDR.String(),
					// The peer connection id is fetched from the cloud formation stack
					// outputs in the stackoutput resource.
					PeerConnectionID: cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
				}

				privateRoutes = append(privateRoutes, route)
			}
		}
	}

	var routeTables *template.ParamsMainRouteTables
	{
		routeTables = &template.ParamsMainRouteTables{
			PrivateRoutes: privateRoutes,
		}
	}

	return routeTables, nil
}

func (r *Resource) newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		recordSets, err := r.newRecordSetsParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		routeTables, err := r.newRouteTablesParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			RecordSets:  recordSets,
			RouteTables: routeTables,
		}
	}

	return params, nil
}

// updateStack is a special implementation of updating the TCCPF stack in the
// when we have to update route tables in case of an upgrade from 1 to 3
// masters. The update is then processed in two update steps. The first stack
// update removes the registered route tables. The second stack update adds the
// new route tables. The reason we cannot update the route tables in place is
// that Cloud Formation is not able to transition properly from current to
// desired state in case the order of Availability Zones we use for route table
// definitions changes. Thus the workaround is to delete and re-create instead
// of update in place.
func (r *Resource) updateStack(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "computing the template of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "removing route tables",
		)

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		params.RouteTables.PrivateRoutes = nil

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "computed the template of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "removing route tables",
		)
	}

	{
		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "requesting the update of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "removing route tables",
		)

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			StackName:    aws.String(key.StackNameTCCPF(&cr)),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.ControlPlane.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "requested the update of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "removing route tables",
		)
	}

	{
		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "waiting for the update of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "removing route tables",
		)

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPF(&cr)),
		}

		err = cc.Client.ControlPlane.AWS.CloudFormation.WaitUntilStackUpdateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "waited for the update of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "removing route tables",
		)
	}

	{
		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "computing the template of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "adding route tables",
		)

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx,
			"level", "debug", "message",
			"computed the template of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "adding route tables",
		)
	}

	{
		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "requesting the update of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "adding route tables",
		)

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			StackName:    aws.String(key.StackNameTCCPF(&cr)),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.ControlPlane.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx,
			"level", "debug",
			"message", "requested the update of the tenant cluster's control plane finalizer cloud formation stack",
			"reason", "adding route tables",
		)
	}

	return nil
}
