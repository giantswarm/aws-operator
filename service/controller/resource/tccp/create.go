package tccp

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	pkgtemplate "github.com/giantswarm/aws-operator/pkg/template"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/ebs"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp/template"
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
		// When a tenant cluster is created, the CPI resource creates a peer role and
		// with it an ARN for it. As long as the peer role ARN is not present, we have
		// to cancel the resource to prevent further TCCP resource actions.
		if cc.Status.ControlPlane.PeerRole.ARN == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane peer role arn is not yet set up")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		// When the TCCP cloud formation stack is transitioning, it means it is
		// updating in most cases. We do not want to interfere with the current
		// process and stop here. We will then check on the next reconciliation loop
		// and continue eventually.
		if cc.Status.TenantCluster.TCCP.IsTransitioning {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane cloud formation stack is in transitioning state")
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

		if len(cc.Status.TenantCluster.TCCP.AvailabilityZones) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "availability zone information not yet available")
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
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", o.Stacks[0].StackStatus)

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
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

	return nil
}

func (r *Resource) createStack(ctx context.Context, cr v1alpha1.Cluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if r.encrypterBackend == encrypter.VaultBackend {
		err = r.encrypterRoleManager.EnsureCreatedAuthorizedIAMRoles(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr, time.Now())
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
	}

	return nil
}

func (r *Resource) detachVolumes(ctx context.Context, cr v1alpha1.Cluster) error {
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

func (r *Resource) getCloudFormationTags(cr v1alpha1.Cluster) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCCP
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) newIAMPoliciesParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var iamPolicies *template.ParamsMainIAMPolicies
	{
		iamPolicies = &template.ParamsMainIAMPolicies{
			ClusterID:         key.ClusterID(&cr),
			EC2ServiceDomain:  key.EC2ServiceDomain(cc.Status.TenantCluster.AWS.Region),
			MasterPolicyName:  key.PolicyNameMaster(cr),
			MasterProfileName: key.ProfileNameMaster(cr),
			MasterRoleName:    key.RoleNameMaster(cr),
			RegionARN:         key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			KMSKeyARN:         cc.Status.TenantCluster.Encryption.Key,
			S3Bucket:          key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
		}
	}

	return iamPolicies, nil
}
func (r *Resource) newInternetGatewayParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainInternetGateway, error) {
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
func (r *Resource) newInstanceParams(ctx context.Context, cr v1alpha1.Cluster, t time.Time) (*template.ParamsMainInstance, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c := template.SmallCloudconfigConfig{
		S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID), key.S3ObjectPathTCCP(&cr)),
	}
	rendered, err := pkgtemplate.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var instance *template.ParamsMainInstance
	{
		instance = &template.ParamsMainInstance{
			Cluster: template.ParamsMainInstanceCluster{
				ID: key.ClusterID(&cr),
			},
			Image: template.ParamsMainInstanceImage{
				ID: key.ImageID(cc.Status.TenantCluster.AWS.Region),
			},
			Master: template.ParamsMainInstanceMaster{
				AZ:               key.MasterAvailabilityZone(cr),
				CloudConfig:      base64.StdEncoding.EncodeToString([]byte(rendered)),
				EncrypterBackend: r.encrypterBackend,
				DockerVolume: template.ParamsMainInstanceMasterDockerVolume{
					Name:         key.VolumeNameDocker(cr),
					ResourceName: key.DockerVolumeResourceName(cr, t),
				},
				EtcdVolume: template.ParamsMainInstanceMasterEtcdVolume{
					Name: key.VolumeNameEtcd(cr),
				},
				LogVolume: template.ParamsMainInstanceMasterLogVolume{
					Name: key.VolumeNameLog(cr),
				},
				Instance: template.ParamsMainInstanceMasterInstance{
					ResourceName: key.MasterInstanceResourceName(cr, t),
					Type:         key.MasterInstanceType(cr),
					Monitoring:   r.instanceMonitoring,
				},
				PrivateSubnet: key.SanitizeCFResourceName(key.PrivateSubnetName(key.MasterAvailabilityZone(cr))),
			},
		}
	}
	return instance, nil
}
func (r *Resource) newLoadBalancersParams(ctx context.Context, cr v1alpha1.Cluster, t time.Time) (*template.ParamsMainLoadBalancers, error) {
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
		if az.Name != key.MasterAvailabilityZone(cr) {
			continue
		}

		privateSubnets = append(privateSubnets, key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name)))
	}

	var loadBalancers *template.ParamsMainLoadBalancers
	{
		loadBalancers = &template.ParamsMainLoadBalancers{
			APIElbHealthCheckTarget: key.HealthCheckTarget(key.KubernetesSecurePort),
			APIElbName:              key.ELBNameAPI(&cr),
			APIInternalElbName:      key.InternalELBNameAPI(&cr),
			APIElbPortsToOpen: []template.ParamsMainLoadBalancersPortPair{
				{
					PortELB:      key.KubernetesSecurePort,
					PortInstance: key.KubernetesSecurePort,
				},
			},
			APIElbScheme:             externalELBScheme,
			APIInternalElbScheme:     internalELBScheme,
			EtcdElbHealthCheckTarget: key.HealthCheckTarget(key.EtcdPort),
			EtcdElbName:              key.ELBNameEtcd(&cr),
			EtcdElbPortsToOpen: []template.ParamsMainLoadBalancersPortPair{
				{
					PortELB:      key.EtcdPort,
					PortInstance: key.EtcdPort,
				},
			},
			EtcdElbScheme:                    internalELBScheme,
			ELBHealthCheckHealthyThreshold:   healthCheckHealthyThreshold,
			ELBHealthCheckInterval:           healthCheckInterval,
			ELBHealthCheckTimeout:            healthCheckTimeout,
			ELBHealthCheckUnhealthyThreshold: healthCheckUnhealthyThreshold,
			MasterInstanceResourceName:       key.MasterInstanceResourceName(cr, t),
			PublicSubnets:                    publicSubnets,
			PrivateSubnets:                   privateSubnets,
		}
	}
	return loadBalancers, nil
}
func (r *Resource) newNATGatewayParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainNATGateway, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var gateways []template.Gateway
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		gw := template.Gateway{
			AvailabilityZone: az.Name,
			NATGWName:        key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATEIPName:       key.SanitizeCFResourceName(key.NATEIPName(az.Name)),
			PublicSubnetName: key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)),
		}
		gateways = append(gateways, gw)
	}

	var natRoutes []template.NATRoute
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cr) {
			continue
		}

		nr := template.NATRoute{
			NATGWName:             key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATRouteName:          key.SanitizeCFResourceName(key.NATRouteName(az.Name)),
			PrivateRouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		natRoutes = append(natRoutes, nr)
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
func (r *Resource) newOutputsParams(ctx context.Context, cr v1alpha1.Cluster, t time.Time) (*template.ParamsMainOutputs, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var outputs *template.ParamsMainOutputs
	{
		outputs = &template.ParamsMainOutputs{
			Master: template.ParamsMainOutputsMaster{
				ImageID: key.ImageID(cc.Status.TenantCluster.AWS.Region),
				Instance: template.ParamsMainOutputsMasterInstance{
					ResourceName: key.MasterInstanceResourceName(cr, t),
					Type:         key.MasterInstanceType(cr),
				},
				DockerVolume: template.ParamsMainOutputsMasterDockerVolume{
					ResourceName: key.DockerVolumeResourceName(cr, t),
				},
			},
			OperatorVersion: key.OperatorVersion(&cr),
			Route53Enabled:  r.route53Enabled,
		}
	}

	return outputs, nil
}
func (r *Resource) newRecordSetsParams(ctx context.Context, cr v1alpha1.Cluster, t time.Time) (*template.ParamsMainRecordSets, error) {
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
func (r *Resource) newRouteTablesParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainRouteTables, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var publicRouteTableNames []template.RouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.RouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		publicRouteTableNames = append(publicRouteTableNames, rtName)
	}

	var privateRouteTableNames []template.RouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cr) {
			continue
		}

		rtName := template.RouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		privateRouteTableNames = append(privateRouteTableNames, rtName)
	}

	var routeTables *template.ParamsMainRouteTables
	{
		routeTables = &template.ParamsMainRouteTables{
			HostClusterCIDR:        cc.Status.ControlPlane.VPC.CIDR,
			PrivateRouteTableNames: privateRouteTableNames,
			PublicRouteTableNames:  publicRouteTableNames,
		}
	}

	return routeTables, nil
}

func getMasterRules(cfg securityConfig, hostClusterCIDR string) ([]template.SecurityGroupRule, error) {
	// Allow traffic to the Kubernetes API server depending on the API
	// whitelisting rules.
	publicAPIRules, err := getKubernetesPublicAPIRules(cfg, hostClusterCIDR)
	if err != nil {
		return []template.SecurityGroupRule{}, microerror.Mask(err)
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

func getEtcdRules(customObject v1alpha1.Cluster, hostClusterCIDR string) []template.SecurityGroupRule {
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

func getKubernetesPrivateAPIRules(cfg securityConfig, hostClusterCIDR string) ([]template.SecurityGroupRule, error) {
	// When public API whitelisting is enabled, add separate security group rule per each subnet.
	if cfg.APIWhitelist.Private.Enabled {
		// Allow control-plane CIDR and tenant cluster CIDR.
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
				SourceCIDR:  key.StatusClusterNetworkCIDR(cfg.CustomObject),
			},
		}

		// Whitelist all configured subnets.
		privateWhitelistSubnets := strings.Split(cfg.APIWhitelist.Private.SubnetList, ",")
		for _, subnet := range privateWhitelistSubnets {
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

		return rules, nil
	} else {
		// When private API whitelisting is disabled, allow all private subnets traffic.
		allowAllRule := []template.SecurityGroupRule{
			{
				Description: "Allow all traffic to the master instance from A class network.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  "10.0.0.0/8",
			},
			{
				Description: "Allow all traffic to the master instance from B class network.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  "172.16.0.0/12",
			},
			{
				Description: "Allow all traffic to the master instance from C class network.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  "192.168.0.0/16",
			},
		}

		return allowAllRule, nil
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
				SourceCIDR:  key.StatusClusterNetworkCIDR(cfg.CustomObject),
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
			return []template.SecurityGroupRule{}, microerror.Mask(err)
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

func (r *Resource) newSecurityGroupsParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainSecurityGroups, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var cfg securityConfig
	{
		cfg = securityConfig{
			APIWhitelist:                    r.apiWhiteList,
			ControlPlaneNATGatewayAddresses: cc.Status.ControlPlane.NATGateway.Addresses,
			ControlPlaneVPCCidr:             cc.Status.ControlPlane.VPC.CIDR,
			CustomObject:                    cr,
		}
	}

	masterRules, err := getMasterRules(cfg, cfg.ControlPlaneVPCCidr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	internalAPIRules, err := getKubernetesPrivateAPIRules(cfg, cfg.ControlPlaneVPCCidr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var securityGroups *template.ParamsMainSecurityGroups
	{
		securityGroups = &template.ParamsMainSecurityGroups{
			APIInternalELBSecurityGroupName:  key.SecurityGroupName(&cfg.CustomObject, "internal-api"),
			APIInternalELBSecurityGroupRules: internalAPIRules,
			APIWhitelistEnabled:              cfg.APIWhitelist.Public.Enabled,
			PrivateAPIWhitelistEnabled:       cfg.APIWhitelist.Private.Enabled,
			MasterSecurityGroupName:          key.SecurityGroupName(&cfg.CustomObject, "master"),
			MasterSecurityGroupRules:         masterRules,
			EtcdELBSecurityGroupName:         key.SecurityGroupName(&cfg.CustomObject, "etcd-elb"),
			EtcdELBSecurityGroupRules:        getEtcdRules(cfg.CustomObject, cfg.ControlPlaneVPCCidr),
		}
	}

	return securityGroups, nil
}
func (r *Resource) newSubnetsParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainSubnets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	// XXX: DEBUG
	fmt.Printf("\n\n================ XXXXXXXXXXXXxx ===========================\nnewSubnetsParams: zones: %#v\n\n\n", zones)

	var publicSubnets []template.Subnet
	for _, az := range zones {
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
		}
		publicSubnets = append(publicSubnets, snet)
	}

	var privateSubnets []template.Subnet
	for _, az := range zones {
		if az.Name != key.MasterAvailabilityZone(cr) {
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
func (r *Resource) newVPCParams(ctx context.Context, cr v1alpha1.Cluster) (*template.ParamsMainVPC, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var routeTableNames []template.RouteTableName
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		rtName := template.RouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}
		routeTableNames = append(routeTableNames, rtName)
	}
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cr) {
			continue
		}

		rtName := template.RouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		routeTableNames = append(routeTableNames, rtName)
	}

	var vpc *template.ParamsMainVPC
	{
		vpc = &template.ParamsMainVPC{
			CidrBlock:        key.StatusClusterNetworkCIDR(cr),
			ClusterID:        key.ClusterID(&cr),
			InstallationName: r.installationName,
			HostAccountID:    cc.Status.ControlPlane.AWSAccountID,
			PeerVPCID:        r.vpcPeerID,
			Region:           key.Region(cr),
			RegionARN:        key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			PeerRoleArn:      cc.Status.ControlPlane.PeerRole.ARN,
			RouteTableNames:  routeTableNames,
		}
	}

	return vpc, nil
}

func (r *Resource) newTemplateParams(ctx context.Context, cr v1alpha1.Cluster, t time.Time) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		iamPolicies, err := r.newIAMPoliciesParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		internetGateway, err := r.newInternetGatewayParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		instance, err := r.newInstanceParams(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		loadBalancers, err := r.newLoadBalancersParams(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		natGateway, err := r.newNATGatewayParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := r.newOutputsParams(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		recordSets, err := r.newRecordSetsParams(ctx, cr, t)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		routeTables, err := r.newRouteTablesParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		securityGroups, err := r.newSecurityGroupsParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		subnets, err := r.newSubnetsParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		vpc, err := r.newVPCParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{

			IAMPolicies:     iamPolicies,
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

func (r *Resource) updateStack(ctx context.Context, cr v1alpha1.Cluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr, time.Now())
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
	}

	return nil
}
