package tccpn

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpn/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
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

	// Ensure some preconditions are met so we have all neccessary information
	// available to manage the TCNP CF stack.
	{
		if cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "EtcdVolumeSnapshotID not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPN(&cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane nodes cloud formation stack")
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
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", o.Stacks[0].StackStatus)

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane nodes cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane nodes cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane nodes cloud formation stack already exists")
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

func (r *Resource) createStack(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := newTemplateParams(ctx, cr, r.apiWhitelist)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane nodes cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCCPN(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane nodes cloud formation stack")
	}

	return nil
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSControlPlane) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCCPN
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) updateStack(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := newTemplateParams(ctx, cr, r.apiWhitelist)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane nodes cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the update of the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			StackName:    aws.String(key.StackNameTCCPN(&cr)),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the update of the tenant cluster's control plane nodes cloud formation stack")
	}

	return nil
}

func newAutoScalingGroup(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainAutoScalingGroup, error) {

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		AvailabilityZone: key.ControlPlaneAvailabilityZones(cr)[0],
		ClusterID:        key.ClusterID(&cr),
		Subnet:           key.SanitizeCFResourceName(key.PrivateSubnetName(key.ControlPlaneAvailabilityZones(cr)[0])),
	}

	return autoScalingGroup, nil
}

func newEtcdVolume(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainEtcdVolume, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	etcdVolume := &template.ParamsMainEtcdVolume{
		AvailabilityZone: key.ControlPlaneAvailabilityZones(cr)[0],
		Name:             key.VolumeNameEtcdCP(cr),
		SnapshotID:       cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID,
	}

	return etcdVolume, nil
}

func newIAMPolicies(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var iamPolicies *template.ParamsMainIAMPolicies
	{
		iamPolicies = &template.ParamsMainIAMPolicies{
			ClusterID:        key.ClusterID(&cr),
			EC2ServiceDomain: key.EC2ServiceDomain(cc.Status.TenantCluster.AWS.Region),
			KMSKeyARN:        cc.Status.TenantCluster.Encryption.Key,
			RegionARN:        key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			S3Bucket:         key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
		}
	}

	return iamPolicies, nil
}

func newLaunchConfiguration(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainLaunchConfiguration, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	launchConfiguration := &template.ParamsMainLaunchConfiguration{
		BlockDeviceMapping: template.ParamsMainLaunchConfigurationBlockDeviceMapping{
			Docker: template.ParamsMainLaunchConfigurationBlockDeviceMappingDocker{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingDockerVolume{
					Size: 100,
				},
			},
			Kubelet: template.ParamsMainLaunchConfigurationBlockDeviceMappingKubelet{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingKubeletVolume{
					Size: 100,
				},
			},
			Logging: template.ParamsMainLaunchConfigurationBlockDeviceMappingLogging{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingLoggingVolume{
					Size: 100,
				},
			},
		},
		Instance: template.ParamsMainLaunchConfigurationInstance{
			Image:      key.ImageID(cc.Status.TenantCluster.AWS.Region),
			Monitoring: true,
			Type:       key.ControlPlaneInstanceType(cr),
		},
		SmallCloudConfig: template.ParamsMainLaunchConfigurationSmallCloudConfig{
			S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID), key.S3ObjectPathTCCPN(&cr)),
		},
	}

	return launchConfiguration, nil
}

func newOutputs(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainOutputs, error) {
	outputs := &template.ParamsMainOutputs{
		InstanceType:    key.ControlPlaneInstanceType(cr),
		OperatorVersion: key.OperatorVersion(&cr),
	}

	return outputs, nil
}

func newRouteTables(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainRouteTables, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var publicRouteTableNames []template.ParamsMainRouteTablesRouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		if az.Name != key.ControlPlaneAvailabilityZones(cr)[0] {
			continue
		}
		rtName := template.ParamsMainRouteTablesRouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		publicRouteTableNames = append(publicRouteTableNames, rtName)
	}

	var privateRouteTableNames []template.ParamsMainRouteTablesRouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		if az.Name != key.ControlPlaneAvailabilityZones(cr)[0] {
			continue
		}

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
			ClusterID:              key.ClusterID(&cr),
			HostClusterCIDR:        cc.Status.ControlPlane.VPC.CIDR,
			PrivateRouteTableNames: privateRouteTableNames,
			PublicRouteTableNames:  publicRouteTableNames,
			VPCID:                  cc.Status.TenantCluster.TCCP.VPC.ID,
			PeeringConnectionID:    cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
		}
	}

	return routeTables, nil
}

func newSecurityGroups(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, apiWhiteList APIWhitelist) (*template.ParamsMainSecurityGroups, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var cfg securityConfig
	{
		cfg = securityConfig{
			APIWhitelist:                    apiWhiteList,
			ControlPlaneNATGatewayAddresses: cc.Status.ControlPlane.NATGateway.Addresses,
			ControlPlaneVPCCidr:             cc.Status.ControlPlane.VPC.CIDR,
			//TODO LH no idea if this is good or not .. was cluster.Status.Provider.Network.CIDR in tccp cluster
			ProviderCIDR: cc.Status.ControlPlane.VPC.CIDR,
			CustomObject: cr,
		}
	}

	masterRules, err := getMasterRules(cfg, cfg.ControlPlaneVPCCidr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var securityGroups *template.ParamsMainSecurityGroups
	{
		securityGroups = &template.ParamsMainSecurityGroups{
			APIWhitelistEnabled:        cfg.APIWhitelist.Public.Enabled,
			PrivateAPIWhitelistEnabled: cfg.APIWhitelist.Private.Enabled,
			MasterSecurityGroupName:    key.SecurityGroupName(&cfg.CustomObject, "master"),
			MasterSecurityGroupRules:   masterRules,
			EtcdELBSecurityGroupName:   key.SecurityGroupName(&cfg.CustomObject, "etcd-elb"),
			EtcdELBSecurityGroupRules:  getEtcdRules(cfg.CustomObject, cfg.ControlPlaneVPCCidr),
			VPCID:                      cc.Status.TenantCluster.TCCP.VPC.ID,
		}
	}

	return securityGroups, nil
}
func newSubnets(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainSubnets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	var publicSubnets []template.Subnet
	for _, az := range zones {
		if az.Name != key.ControlPlaneAvailabilityZones(cr)[0] {
			continue
		}

		snetName := key.SanitizeCFResourceName(key.PublicSubnetName(az.Name))
		snet := template.Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR.String(),
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: template.RouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PublicSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
			VPCID: cc.Status.TenantCluster.TCCP.VPC.ID,
		}
		publicSubnets = append(publicSubnets, snet)
	}

	var privateSubnets []template.Subnet
	for _, az := range zones {
		if az.Name != key.ControlPlaneAvailabilityZones(cr)[0] {
			continue
		}

		snetName := key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name))
		snet := template.Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR.String(),
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: template.RouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PrivateSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
			VPCID: cc.Status.TenantCluster.TCCP.VPC.ID,
		}
		privateSubnets = append(privateSubnets, snet)
	}

	var subnets *template.ParamsMainSubnets
	{
		subnets = &template.ParamsMainSubnets{
			PublicSubnets:  publicSubnets,
			PrivateSubnets: privateSubnets,
		}
	}

	return subnets, nil
}

func newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, apiWhiteList APIWhitelist) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		autoScalingGroup, err := newAutoScalingGroup(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		etcdVolume, err := newEtcdVolume(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		iamPolicies, err := newIAMPolicies(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		launchConfiguration, err := newLaunchConfiguration(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := newOutputs(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		routeTables, err := newRouteTables(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		securityGroups, err := newSecurityGroups(ctx, cr, apiWhiteList)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		subnets, err := newSubnets(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			AutoScalingGroup:    autoScalingGroup,
			EtcdVolume:          etcdVolume,
			IAMPolicies:         iamPolicies,
			LaunchConfiguration: launchConfiguration,
			Outputs:             outputs,
			RouteTables:         routeTables,
			SecurityGroups:      securityGroups,
			Subnets:             subnets,
		}
	}

	return params, nil
}

func getMasterRules(cfg securityConfig, hostClusterCIDR string) ([]template.SecurityGroupRule, error) {
	// Allow traffic to the Kubernetes API server depending on the API
	// whitelisting rules.
	publicAPIRules, err := getKubernetesPublicAPIRules(cfg, hostClusterCIDR)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Other security group rules for the master.
	otherRules := []template.SecurityGroupRule{
		{
			Description: "Allow traffic from control plane CIDR to 4194 for cadvisor scraping.",
			Port:        cadvisorPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 2379 for etcd backup.",
			Port:        etcdPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 10250 for kubelet scraping.",
			Port:        kubeletPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 10300 for node-exporter scraping.",
			Port:        nodeExporterPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 10301 for kube-state-metrics scraping.",
			Port:        kubeStateMetricsPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Only allow ssh traffic from the control plane.",
			Port:        sshPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
	}

	return append(publicAPIRules, otherRules...), nil
}

func getEtcdRules(customObject infrastructurev1alpha2.AWSControlPlane, hostClusterCIDR string) []template.SecurityGroupRule {
	return []template.SecurityGroupRule{
		{
			Description: "Allow all etcd traffic from the VPC to the etcd load balancer.",
			Port:        etcdPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  defaultCIDR,
		},
		{
			Description: "Allow traffic from control plane to etcd port for backup and metrics.",
			Port:        etcdPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
	}
}

func getKubernetesPublicAPIRules(cfg securityConfig, hostClusterCIDR string) ([]template.SecurityGroupRule, error) {
	// When API whitelisting is enabled, add separate security group rule per each subnet.
	if cfg.APIWhitelist.Public.Enabled {
		rules := []template.SecurityGroupRule{
			{
				Description: "Allow traffic from control plane CIDR.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  hostClusterCIDR,
			},
			{
				Description: "Allow traffic from tenant cluster CIDR.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				//TODO LH what value needs to go in here for tccpn ?
				SourceCIDR: cfg.ProviderCIDR,
			},
		}

		// Whitelist all configured subnets.
		publicWhitelistSubnets := strings.Split(cfg.APIWhitelist.Public.SubnetList, ",")
		for _, subnet := range publicWhitelistSubnets {
			if subnet != "" {
				subnetRule := template.SecurityGroupRule{
					Description: "Custom Whitelist CIDR.",
					Port:        key.KubernetesSecurePort,
					Protocol:    tcpProtocol,
					SourceCIDR:  subnet,
				}
				rules = append(rules, subnetRule)
			}
		}

		// Whitelist public EIPs of the host cluster NAT gateways.
		hostClusterNATGatewayRules, err := getHostClusterNATGatewayRules(cfg)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, gatewayRule := range hostClusterNATGatewayRules {
			rules = append(rules, gatewayRule)
		}

		return rules, nil
	} else {
		// When API whitelisting is disabled, allow all traffic.
		allowAllRule := []template.SecurityGroupRule{
			{
				Description: "Allow all traffic to the master instance.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  defaultCIDR,
			},
		}

		return allowAllRule, nil
	}
}

func getHostClusterNATGatewayRules(cfg securityConfig) ([]template.SecurityGroupRule, error) {
	var gatewayRules []template.SecurityGroupRule

	for _, address := range cfg.ControlPlaneNATGatewayAddresses {
		gatewayRule := template.SecurityGroupRule{
			Description: "Allow traffic from gateways.",
			Port:        key.KubernetesSecurePort,
			Protocol:    tcpProtocol,
			SourceCIDR:  fmt.Sprintf("%s/32", *address.PublicIp),
		}

		gatewayRules = append(gatewayRules, gatewayRule)
	}

	return gatewayRules, nil
}
