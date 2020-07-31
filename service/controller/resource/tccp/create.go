package tccp

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp/template"
	"github.com/giantswarm/aws-operator/service/internal/ebs"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
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
		// When aws operator starts it needs to find CP VPC information, so we have to
		// cancel the resource in case the information is not available yet.
		if cc.Status.ControlPlane.VPC.ID == "" || cc.Status.ControlPlane.VPC.CIDR == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "control plane VPC info not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		// When a tenant cluster is created, the CPI resource creates a peer role and
		// with it an ARN for it. As long as the peer role ARN is not present, we have
		// to cancel the resource to prevent further TCCP resource actions.
		if cc.Status.ControlPlane.PeerRole.ARN == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster's control plane peer role arn not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		// When the TCCP cloud formation stack is transitioning, it means it is
		// updating in most cases. We do not want to interfere with the current
		// process and stop here. We will then check on the next reconciliation loop
		// and continue eventually.
		if cc.Status.TenantCluster.TCCP.IsTransitioning {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster's control plane cloud formation stack is in transitioning state")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		// The IPAM resource is executed before the CloudFormation resource in order
		// to allocate a free IP range for the tenant subnet. This CIDR is put into
		// the CR status. In case it is missing, the IPAM resource did not yet
		// allocate it and the CloudFormation resource cannot proceed. We cancel here
		// and wait for the CIDR to be available in the CR status.
		if key.StatusClusterNetworkCIDR(cr) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane network cidr")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCP(&cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane cloud formation stack")

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
			return microerror.Maskf(executionCFCreateError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusRollbackFailed {
			return microerror.Maskf(executionCFRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateRollbackFailed {
			return microerror.Maskf(executionCFUpdateRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.event.Emit(ctx, &cr, "CFCreate", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateInProgress))
			r.event.Emit(ctx, &cr, "CFUpdate", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusRollbackInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusRollbackInProgress))
			r.event.Emit(ctx, &cr, "CFRollback", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusRollbackInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateRollbackInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusRollbackInProgress))
			r.event.Emit(ctx, &cr, "CFUpdateRollback", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateRollbackInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateComplete {
			r.event.Emit(ctx, &cr, "CFCreated", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusCreateComplete))
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateComplete {
			r.event.Emit(ctx, &cr, "CFUpdated", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateComplete))
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusRollbackComplete {
			r.event.Emit(ctx, &cr, "CFRollbackCompleted", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusRollbackComplete))
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateRollbackComplete {
			r.event.Emit(ctx, &cr, "CFUpdateRollbackCompleted", fmt.Sprintf("The tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateRollbackComplete))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane cloud formation stack")
	}

	{
		update, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if update {
			err = r.stopMasterInstance(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.detachVolumes(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.snapshotEtcdVolume(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.terminateMasterInstance(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.updateStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	// TODO remove etcd snapshot migration code
	// https://github.com/giantswarm/giantswarm/issues/9979
	if key.IsNewCluster(cr) && cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID == "" {
		err = r.snapshotEtcdVolume(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created etcd volume snapshot for newly created cluster")
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane cloud formation stack")

		params, err := r.newParamsMain(ctx, cr, time.Now())
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(namedIAMCapability),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCCP(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane cloud formation stack")
		r.event.Emit(ctx, &cr, "CFCreateRequested", "Requested the creation of the tenant cluster's control plane cloud formation stack")
	}

	return nil
}

func (r *Resource) detachVolumes(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var ebsService ebs.Interface
	{
		c := ebs.Config{
			Client: cc.Client.TenantCluster.AWS.EC2,
			Logger: r.logger,
		}

		ebsService, err = ebs.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		// Fetch the etcd volume information.
		filterFuncs := []func(t *ec2.Tag) bool{
			ebs.NewDockerVolumeFilter(cr),
			ebs.NewEtcdVolumeFilter(cr),
		}
		volumes, err := ebsService.ListVolumes(cr, filterFuncs...)
		if err != nil {
			return microerror.Mask(err)
		}

		force := false
		shutdown := false
		wait := false

		for _, v := range volumes {
			for _, a := range v.Attachments {
				err := ebsService.DetachVolume(ctx, v.VolumeID, a, force, shutdown, wait)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}

	return nil
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSCluster) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCCP
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) newParamsMain(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, t time.Time) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		internetGateway, err := r.newParamsMainInternetGateway(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		instance, err := r.newParamsMainInstance(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		loadBalancers, err := r.newParamsMainLoadBalancers(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		natGateway, err := r.newParamsMainNATGateway(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := r.newParamsMainOutputs(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		recordSets, err := r.newParamsMainRecordSets(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		routeTables, err := r.newParamsMainRouteTables(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		securityGroups, err := r.newParamsMainSecurityGroups(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		subnets, err := r.newParamsMainSubnets(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		vpc, err := r.newParamsMainVPC(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			InternetGateway: internetGateway,
			Instance:        instance,
			LoadBalancers:   loadBalancers,
			NATGateway:      natGateway,
			Outputs:         outputs,
			RecordSets:      recordSets,
			RouteTables:     routeTables,
			SecurityGroups:  securityGroups,
			Subnets:         subnets,
			VPC:             vpc,
		}
	}

	return params, nil
}

func (r *Resource) newParamsMainInstance(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, t time.Time) (*template.ParamsMainInstance, error) {
	var err error
	// We need to fetch the ControlPlane CR for once because it needs to be passed to hamaster interface to get the mapopings.
	var haMapping []hamaster.Mapping
	{
		haMapping, err = r.haMaster.Mapping(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var instance *template.ParamsMainInstance
	{
		instance = &template.ParamsMainInstance{
			Master: template.ParamsMainInstanceMaster{
				/*
					This is part of migration code we introduced with the new tccpn stack.
					We need this to create empty ETCD volume in the first AZ  to do a snapshot
					and then create volume from that snapshot in the tccpn stack.

					We can take a look at how we could remove it before creating stable release for HA masters.
				*/
				AZ: haMapping[0].AZ,
				EtcdVolume: template.ParamsMainInstanceMasterEtcdVolume{
					Name: key.VolumeNameEtcd(cr),
				},
			},
		}
	}
	return instance, nil
}

func (r *Resource) newParamsMainInternetGateway(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainInternetGateway, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var internetGateways []template.ParamsMainInternetGatewayInternetGateway
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		ig := template.ParamsMainInternetGatewayInternetGateway{
			InternetGatewayRoute: key.SanitizeCFResourceName(key.PublicInternetGatewayRouteName(az.Name)),
			RouteTable:           key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}

		internetGateways = append(internetGateways, ig)
	}

	var internetGateway *template.ParamsMainInternetGateway
	{
		internetGateway = &template.ParamsMainInternetGateway{
			ClusterID:        key.ClusterID(&cr),
			InternetGateways: internetGateways,
		}
	}

	return internetGateway, nil
}

func (r *Resource) newParamsMainLoadBalancers(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, t time.Time) (*template.ParamsMainLoadBalancers, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)

	}

	clusterAZs := cc.Spec.TenantCluster.TCCP.AvailabilityZones
	if len(clusterAZs) < 1 {
		return nil, microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
	}

	var publicSubnets []string
	for _, az := range clusterAZs {
		publicSubnets = append(publicSubnets, key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)))
	}

	var privateSubnets []string
	for _, az := range clusterAZs {
		privateSubnets = append(privateSubnets, key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name)))
	}

	var loadBalancers *template.ParamsMainLoadBalancers
	{
		loadBalancers = &template.ParamsMainLoadBalancers{
			APIElbHealthCheckTarget: key.HealthCheckHTTPTarget(key.KubernetesApiHealthCheckPort),
			APIElbName:              key.ELBNameAPI(&cr),
			APIInternalElbName:      key.InternalELBNameAPI(&cr),
			APIElbPortsToOpen: []template.ParamsMainLoadBalancersPortPair{
				{
					PortELB:      key.KubernetesSecurePort,
					PortInstance: key.KubernetesSecurePort,
				},
			},
			EtcdElbHealthCheckTarget: key.HealthCheckTCPTarget(key.EtcdPort),
			EtcdElbName:              key.ELBNameEtcd(&cr),
			EtcdElbPortsToOpen: []template.ParamsMainLoadBalancersPortPair{
				{
					PortELB:      key.EtcdPort,
					PortInstance: key.EtcdPort,
				},
			},
			MasterInstanceResourceName: key.MasterInstanceResourceName(cr, t),
			PublicSubnets:              publicSubnets,
			PrivateSubnets:             privateSubnets,
		}
	}

	return loadBalancers, nil
}

func (r *Resource) newParamsMainNATGateway(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainNATGateway, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var gateways []template.ParamsMainNATGatewayGateway
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		gw := template.ParamsMainNATGatewayGateway{
			AvailabilityZone: az.Name,
			ClusterID:        key.ClusterID(&cr),
			NATGWName:        key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATEIPName:       key.SanitizeCFResourceName(key.NATEIPName(az.Name)),
			PublicSubnetName: key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)),
		}
		gateways = append(gateways, gw)
	}

	var natRoutes []template.ParamsMainNATGatewayNATRoute
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		{
			nr := template.ParamsMainNATGatewayNATRoute{
				NATGWName:      key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
				NATRouteName:   key.SanitizeCFResourceName(key.NATRouteName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			}

			natRoutes = append(natRoutes, nr)
		}

		{
			nr := template.ParamsMainNATGatewayNATRoute{
				NATGWName:      key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
				NATRouteName:   key.SanitizeCFResourceName(key.AWSCNINATRouteName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.AWSCNIRouteTableName(az.Name)),
			}

			natRoutes = append(natRoutes, nr)
		}

	}

	var natGateway *template.ParamsMainNATGateway
	{
		natGateway = &template.ParamsMainNATGateway{
			Gateways:  gateways,
			NATRoutes: natRoutes,
		}
	}

	return natGateway, nil
}

func (r *Resource) newParamsMainOutputs(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, t time.Time) (*template.ParamsMainOutputs, error) {
	var outputs *template.ParamsMainOutputs
	{
		outputs = &template.ParamsMainOutputs{
			OperatorVersion: key.OperatorVersion(&cr),
			Route53Enabled:  r.route53Enabled,
		}
	}

	return outputs, nil
}

func (r *Resource) newParamsMainRecordSets(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, t time.Time) (*template.ParamsMainRecordSets, error) {
	_, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var recordSets *template.ParamsMainRecordSets
	{
		recordSets = &template.ParamsMainRecordSets{
			BaseDomain:                 key.ClusterBaseDomain(cr),
			EtcdDomain:                 key.ClusterEtcdEndpoint(cr),
			ClusterID:                  key.ClusterID(&cr),
			MasterInstanceResourceName: key.MasterInstanceResourceName(cr, t),
			Route53Enabled:             r.route53Enabled,
			VPCRegion:                  key.Region(cr),
		}
	}

	return recordSets, nil
}

func (r *Resource) newParamsMainRouteTables(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainRouteTables, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var awsCNIRouteTableNames []template.ParamsMainRouteTablesRouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.ParamsMainRouteTablesRouteTableName{
			AvailabilityZone: az.Name,
			ResourceName:     key.SanitizeCFResourceName(key.AWSCNIRouteTableName(az.Name)),
		}
		awsCNIRouteTableNames = append(awsCNIRouteTableNames, rtName)
	}

	var publicRouteTableNames []template.ParamsMainRouteTablesRouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.ParamsMainRouteTablesRouteTableName{
			AvailabilityZone: az.Name,
			ResourceName:     key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}
		publicRouteTableNames = append(publicRouteTableNames, rtName)
	}

	var privateRouteTableNames []template.ParamsMainRouteTablesRouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.ParamsMainRouteTablesRouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		privateRouteTableNames = append(privateRouteTableNames, rtName)
	}

	var routeTables *template.ParamsMainRouteTables
	{
		routeTables = &template.ParamsMainRouteTables{
			ClusterID:       key.ClusterID(&cr),
			HostClusterCIDR: cc.Status.ControlPlane.VPC.CIDR,

			AWSCNIRouteTableNames:  awsCNIRouteTableNames,
			PrivateRouteTableNames: privateRouteTableNames,
			PublicRouteTableNames:  publicRouteTableNames,
		}
	}

	return routeTables, nil
}

func (r *Resource) newParamsMainSecurityGroups(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainSecurityGroups, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var securityGroups *template.ParamsMainSecurityGroups
	{
		securityGroups = &template.ParamsMainSecurityGroups{
			APIWhitelist: template.ParamsMainSecurityGroupsAPIWhitelist{
				Private: template.ParamsMainSecurityGroupsAPIWhitelistSecurityGroup{
					Enabled:    r.apiWhitelist.Private.Enabled,
					SubnetList: r.apiWhitelist.Private.SubnetList,
				},
				Public: template.ParamsMainSecurityGroupsAPIWhitelistSecurityGroup{
					Enabled:    r.apiWhitelist.Public.Enabled,
					SubnetList: r.apiWhitelist.Public.SubnetList,
				},
			},
			ClusterID:                       key.ClusterID(&cr),
			ControlPlaneNATGatewayAddresses: cc.Status.ControlPlane.NATGateway.Addresses,
			ControlPlaneVPCCIDR:             cc.Status.ControlPlane.VPC.CIDR,
			TenantClusterVPCCIDR:            key.StatusClusterNetworkCIDR(cr),
		}
	}

	return securityGroups, nil
}

func (r *Resource) newParamsMainSubnets(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainSubnets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	var awsCNISubnets []template.ParamsMainSubnetsSubnet
	for _, az := range zones {
		snetName := key.SanitizeCFResourceName(key.AWSCNISubnetName(az.Name))
		snet := template.ParamsMainSubnetsSubnet{
			AvailabilityZone: az.Name,
			CIDR:             az.Subnet.AWSCNI.CIDR.String(),
			Name:             snetName,
			RouteTableAssociation: template.ParamsMainSubnetsSubnetRouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.AWSCNISubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.AWSCNIRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
		}
		awsCNISubnets = append(awsCNISubnets, snet)
	}

	var publicSubnets []template.ParamsMainSubnetsSubnet
	for _, az := range zones {
		snetName := key.SanitizeCFResourceName(key.PublicSubnetName(az.Name))
		snet := template.ParamsMainSubnetsSubnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR.String(),
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: template.ParamsMainSubnetsSubnetRouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PublicSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
		}
		publicSubnets = append(publicSubnets, snet)
	}

	var privateSubnets []template.ParamsMainSubnetsSubnet
	for _, az := range zones {
		snetName := key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name))
		snet := template.ParamsMainSubnetsSubnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR.String(),
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: template.ParamsMainSubnetsSubnetRouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PrivateSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
		}
		privateSubnets = append(privateSubnets, snet)
	}

	var subnets *template.ParamsMainSubnets
	{
		subnets = &template.ParamsMainSubnets{
			AWSCNISubnets:  awsCNISubnets,
			PublicSubnets:  publicSubnets,
			PrivateSubnets: privateSubnets,
		}
	}

	return subnets, nil
}

func (r *Resource) newParamsMainVPC(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (*template.ParamsMainVPC, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var routeTableNames []template.ParamsMainVPCRouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.ParamsMainVPCRouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}
		routeTableNames = append(routeTableNames, rtName)
	}
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.ParamsMainVPCRouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		routeTableNames = append(routeTableNames, rtName)
	}

	// Allow the actual VPC subnet CIDR to be overwritten by the CR spec.
	podSubnet := r.cidrBlockAWSCNI
	if cr.Spec.Provider.Pods.CIDRBlock != "" {
		podSubnet = cr.Spec.Provider.Pods.CIDRBlock
	}

	var vpc *template.ParamsMainVPC
	{
		vpc = &template.ParamsMainVPC{
			CidrBlock:        key.StatusClusterNetworkCIDR(cr),
			CIDRBlockAWSCNI:  podSubnet,
			ClusterID:        key.ClusterID(&cr),
			InstallationName: r.installationName,
			HostAccountID:    cc.Status.ControlPlane.AWSAccountID,
			PeerVPCID:        cc.Status.ControlPlane.VPC.ID,
			Region:           key.Region(cr),
			RegionARN:        key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			PeerRoleArn:      cc.Status.ControlPlane.PeerRole.ARN,
			RouteTableNames:  routeTableNames,
		}
	}

	return vpc, nil
}

func (r *Resource) snapshotEtcdVolume(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID != "" {
		// In case there is a snapshot ID, we already created the snapshot and do
		// not need to do it again.
		r.logger.LogCtx(ctx, "level", "debug", "message", "not creating etcd volume snapshot")
		r.logger.LogCtx(ctx, "level", "debug", "message", "etcd volume snapshot already created")
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating etcd volume snapshot")

	var ebsService ebs.Interface
	{
		c := ebs.Config{
			Client: cc.Client.TenantCluster.AWS.EC2,
			Logger: r.logger,
		}

		ebsService, err = ebs.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var etcdVolumeID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding etcd volume ID")

		volumes, err := ebsService.ListVolumes(cr, ebs.NewEtcdVolumeFilter(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		if len(volumes) != 1 {
			return microerror.Maskf(executionFailedError, "1 etcd volume must exist, got %d", len(volumes))
		}

		etcdVolumeID = volumes[0].VolumeID

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found etcd volume ID %#q", etcdVolumeID))
	}

	{
		i := &ec2.CreateSnapshotInput{
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("snapshot"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(key.TagSnapshot),
							Value: aws.String(key.HAMasterSnapshotIDValue),
						},
						{
							Key:   aws.String(key.TagCluster),
							Value: aws.String(key.ClusterID(&cr)),
						},
					},
				},
			},
			VolumeId: aws.String(etcdVolumeID),
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.CreateSnapshot(i)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created etcd volume snapshot")

	return nil
}

func (r *Resource) updateStack(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane cloud formation stack")

		params, err := r.newParamsMain(ctx, cr, time.Now())
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the update of the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(namedIAMCapability),
			},
			StackName:    aws.String(key.StackNameTCCP(&cr)),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the update of the tenant cluster's control plane cloud formation stack")
		r.event.Emit(ctx, &cr, "CFUpdateRequested", "Requested the update of the tenant cluster's control plane cloud formation stack")
	}

	return nil
}
