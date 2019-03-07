package cpf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
	"github.com/giantswarm/aws-operator/service/controller/v24/templates"
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

	{
		if cc.Status.Cluster.VPCPeeringConnectionID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the VPC Peering Connection ID in the controller context")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainHostPostStackName(cr)),
		}

		_, err = r.cloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
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

		privateRoutes, err := r.newPrivateRoutes(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		publicRoutes, err := r.newPublicRoutes(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		var cpf *adapter.CPF
		{
			c := adapter.CPFConfig{
				BaseDomain:                 key.BaseDomain(cr),
				ClusterID:                  key.ClusterID(cr),
				GuestHostedZoneNameServers: cc.Status.Cluster.HostedZoneNameServers,
				PrivateRoutes:              privateRoutes,
				PublicRoutes:               publicRoutes,
				Route53Enabled:             r.route53Enabled,
			}

			cpf, err = adapter.NewCPF(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		templateBody, err = templates.Render(key.CloudFormationHostPostTemplates(), cpf)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.CreateStackInput{
			EnableTerminationProtection: aws.Bool(key.EnableTerminationProtection),
			StackName:                   aws.String(key.MainHostPostStackName(cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = r.cloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane finalizer cloud formation stack")
	}

	return nil
}

func (r *Resource) newPrivateRoutes(ctx context.Context, cr v1alpha1.AWSConfig) ([]adapter.CPFRouteTablesRoute, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var tenantPrivateSubnetCidrs []string
	{
		for _, az := range key.StatusAvailabilityZones(cr) {
			tenantPrivateSubnetCidrs = append(tenantPrivateSubnetCidrs, az.Subnet.Private.CIDR)
		}
	}

	var routes []adapter.CPFRouteTablesRoute
	for _, name := range r.routeTables {
		id, err := r.routeTable.IDForName(ctx, name)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, cidrBlock := range tenantPrivateSubnetCidrs {
			route := adapter.CPFRouteTablesRoute{
				RouteTableName: name,
				RouteTableID:   id,
				// Requester CIDR block, we create the peering connection from the
				// tenant's private subnets.
				CidrBlock: cidrBlock,
				// The peer connection id is fetched from the cloud formation stack
				// outputs in the stackoutput resource.
				PeerConnectionID: cc.Status.Cluster.VPCPeeringConnectionID,
			}

			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (r *Resource) newPublicRoutes(ctx context.Context, cr v1alpha1.AWSConfig) ([]adapter.CPFRouteTablesRoute, error) {
	if r.encrypterBackend != encrypter.VaultBackend {
		return nil, nil
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var routes []adapter.CPFRouteTablesRoute

	for _, name := range r.routeTables {
		id, err := r.routeTable.IDForName(ctx, name)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		route := adapter.CPFRouteTablesRoute{
			RouteTableName: name,
			RouteTableID:   id,
			// Requester CIDR block, we create the peering connection from the
			// tenant's CIDR for being able to access Vault's ELB.
			CidrBlock: key.StatusNetworkCIDR(cr),
			// The peer connection id is fetched from the cloud formation stack
			// outputs in the stackoutput resource.
			PeerConnectionID: cc.Status.Cluster.VPCPeeringConnectionID,
		}

		routes = append(routes, route)
	}

	return routes, nil
}
