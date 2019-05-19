package cpf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/resource/cpf/template"
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
			StackName: aws.String(key.StackNameCPF(cr)),
		}

		o, err := cc.Client.ControlPlane.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			// fall through

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", o.Stacks[0].StackStatus)

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

		var params *template.ParamsMain
		{
			recordSets, err := r.newRecordSetsParams(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			routeTables, err := r.newRouteTablesParams(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			params = &template.ParamsMain{
				RecordSets:  recordSets,
				RouteTables: routeTables,
			}
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
			StackName:                   aws.String(key.StackNameCPF(cr)),
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
			StackName: aws.String(key.StackNameCPF(cr)),
		}

		err = cc.Client.ControlPlane.AWS.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "waited for the creation of the tenant cluster's control plane finalizer cloud formation stack")
	}

	return nil
}

func (r *Resource) newPrivateRoutes(ctx context.Context, cr v1alpha1.Cluster) ([]template.ParamsMainRouteTablesRoute, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var tenantPrivateSubnetCidrs []string
	{
		for _, az := range key.StatusAvailabilityZones(cc.Status.TenantCluster.TCCP.MachineDeployment) {
			tenantPrivateSubnetCidrs = append(tenantPrivateSubnetCidrs, az.Subnet.Private.CIDR)
		}
	}

	var routes []template.ParamsMainRouteTablesRoute

	for _, id := range cc.Status.ControlPlane.RouteTable.Mappings {
		for _, cidrBlock := range tenantPrivateSubnetCidrs {
			route := template.ParamsMainRouteTablesRoute{
				RouteTableID: id,
				// Requester CIDR block, we create the peering connection from the
				// tenant's private subnets.
				CidrBlock: cidrBlock,
				// The peer connection id is fetched from the cloud formation stack
				// outputs in the stackoutput resource.
				PeerConnectionID: cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
			}

			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (r *Resource) newPublicRoutes(ctx context.Context, cr v1alpha1.Cluster) ([]template.ParamsMainRouteTablesRoute, error) {
	if r.encrypterBackend != encrypter.VaultBackend {
		return nil, nil
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var routes []template.ParamsMainRouteTablesRoute

	for _, id := range cc.Status.ControlPlane.RouteTable.Mappings {
		route := template.ParamsMainRouteTablesRoute{
			RouteTableID: id,
			// Requester CIDR block, we create the peering connection from the
			// tenant's CIDR for being able to access Vault's ELB.
			CidrBlock: key.StatusClusterNetworkCIDR(cr),
			// The peer connection id is fetched from the cloud formation stack
			// outputs in the stackoutput resource.
			PeerConnectionID: cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func (r *Resource) newRecordSetsParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainRecordSets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var recordSets *template.ParamsMainRecordSets
	{
		recordSets = &template.ParamsMainRecordSets{
			BaseDomain:                 key.ClusterBaseDomain(cr),
			ClusterID:                  key.ClusterID(cr),
			GuestHostedZoneNameServers: cc.Status.TenantCluster.HostedZoneNameServers,
			Route53Enabled:             r.route53Enabled,
		}
	}

	return recordSets, nil
}

func (r *Resource) newRouteTablesParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainRouteTables, error) {
	privateRoutes, err := r.newPrivateRoutes(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	publicRoutes, err := r.newPublicRoutes(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
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
