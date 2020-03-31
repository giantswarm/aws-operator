package tccpf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpf/template"
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
			// fall through

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane finalizer cloud formation stack already exists")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane finalizer cloud formation stack")
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane finalizer cloud formation stack")

		params, err := newTemplateParams(ctx, cr, r.encrypterBackend, r.route53Enabled)
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

func newRecordSetsParams(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, route53Enabled bool) (*template.ParamsMainRecordSets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var recordSets *template.ParamsMainRecordSets
	{
		recordSets = &template.ParamsMainRecordSets{
			BaseDomain:                 key.ClusterBaseDomain(cr),
			ClusterID:                  key.ClusterID(&cr),
			GuestHostedZoneNameServers: cc.Status.TenantCluster.HostedZoneNameServers,
			Route53Enabled:             route53Enabled,
		}
	}

	return recordSets, nil
}

func newRouteTablesParams(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, encrypterBackend string) (*template.ParamsMainRouteTables, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var privateRoutes []template.ParamsMainRouteTablesRoute
	{
		for _, rt := range cc.Status.ControlPlane.RouteTables {
			for _, az := range cc.Status.TenantCluster.TCCP.AvailabilityZones {
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

	var publicRoutes []template.ParamsMainRouteTablesRoute
	if encrypterBackend == encrypter.VaultBackend {
		for _, rt := range cc.Status.ControlPlane.RouteTables {
			route := template.ParamsMainRouteTablesRoute{
				RouteTableID: *rt.RouteTableId,
				// Requester CIDR block, we create the peering connection from the
				// tenant's CIDR for being able to access Vault's ELB.
				CidrBlock: key.StatusClusterNetworkCIDR(cr),
				// The peer connection id is fetched from the cloud formation stack
				// outputs in the stackoutput resource.
				PeerConnectionID: cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
			}

			publicRoutes = append(publicRoutes, route)
		}
	}

	var routeTables *template.ParamsMainRouteTables
	{
		routeTables = &template.ParamsMainRouteTables{
			PrivateRoutes: privateRoutes,
			PublicRoutes:  publicRoutes,
		}
	}

	return routeTables, nil
}

func newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, encrypterBackend string, route53Enabled bool) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		recordSets, err := newRecordSetsParams(ctx, cr, route53Enabled)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		routeTables, err := newRouteTablesParams(ctx, cr, encrypterBackend)
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
