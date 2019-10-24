// Package adapter contains the required logic for creating data structures used for
// feeding CloudFormation templates.
//
// It follows the adapter pattern https://en.wikipedia.org/wiki/Adapter_pattern in the
// sense that it has the knowledge to transform a aws custom object into a data structure
// easily interpolable into the templates without any additional view logic.
//
// There's a base template in `service/templates/cloudformation/guest/main.yaml` which defines
// the basic structure and includes the rest of templates that form the stack as nested
// templates. Those subtemplates should use a `define` action with the name that will be
// used to refer to them from the main template, as explained here
// https://golang.org/pkg/text/template/#hdr-Nested_template_definitions
//
// Each adapter is related to one of these nested templates. It includes the data structure
// with all the values needed to interpolate in the related template and the logic required
// to obtain them, this logic is packed into functions called `hydraters`.
//
// When extending the stack we will just need to:
// * Add the template file in `service/template/cloudformation/guest` and modify
// `service/template/cloudformation/main.yaml` to include the new template.
// * Add the adapter logic file in `service/resource/cloudformation/adapter` with the type
// definition and the Hydrater function to fill the fields (like asg.go or
// launch_configuration.go).
// * Add the new type to the Adapter type in `service/resource/cloudformation/adapter/adapter.go`
// and include the Hydrater function in the `hydraters` slice.
package template

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/template"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

type Config struct {
	APIWhitelist                    APIWhitelist
	AWSRegion                       string
	ControlPlaneAccountID           string
	ControlPlaneNATGatewayAddresses []*ec2.Address
	ControlPlanePeerRoleARN         string
	ControlPlaneVPCID               string
	ControlPlaneVPCCidr             string
	CustomObject                    v1alpha1.Cluster
	EncrypterBackend                string
	GuestAccountID                  string
	InstallationName                string
	PublicRouteTables               string
	Route53Enabled                  bool
	StackState                      StackState
	TenantClusterAccountID          string
	TenantClusterKMSKeyARN          string
	TenantClusterAvailabilityZones  []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone
}

type Adapter struct {
	Guest GuestAdapter
}

func NewGuest(cfg Config) (Adapter, error) {
	a := Adapter{}

	hydraters := []Hydrater{
		a.Guest.IAMPolicies.Adapt,
		a.Guest.InternetGateway.Adapt,
		a.Guest.Instance.Adapt,
		a.Guest.LoadBalancers.Adapt,
		a.Guest.NATGateway.Adapt,
		a.Guest.Outputs.Adapt,
		a.Guest.RecordSets.Adapt,
		a.Guest.RouteTables.Adapt,
		a.Guest.SecurityGroups.Adapt,
		a.Guest.Subnets.Adapt,
		a.Guest.VPC.Adapt,
	}

	for _, h := range hydraters {
		if err := h(cfg); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}

func (i *GuestIAMPoliciesAdapter) Adapt(cfg Config) error {
	clusterID := key.ClusterID(&cfg.CustomObject)

	i.ClusterID = clusterID
	i.EC2ServiceDomain = key.EC2ServiceDomain(cfg.AWSRegion)
	i.MasterPolicyName = key.PolicyNameMaster(cfg.CustomObject)
	i.MasterProfileName = key.ProfileNameMaster(cfg.CustomObject)
	i.MasterRoleName = key.RoleNameMaster(cfg.CustomObject)
	i.RegionARN = key.RegionARN(cfg.AWSRegion)
	i.KMSKeyARN = cfg.TenantClusterKMSKeyARN
	i.S3Bucket = key.BucketName(&cfg.CustomObject, cfg.TenantClusterAccountID)

	return nil
}

func (i *GuestInstanceAdapter) Adapt(config Config) error {
	{
		i.Cluster.ID = key.ClusterID(&config.CustomObject)
	}

	{
		i.Image.ID = config.StackState.MasterImageID
	}

	{
		zones := config.TenantClusterAvailabilityZones

		sort.Slice(zones, func(i, j int) bool {
			return zones[i].Name < zones[j].Name
		})

		if len(zones) < 1 {
			return microerror.Maskf(notFoundError, "CustomObject has no availability zones")
		}

		i.Master.AZ = key.MasterAvailabilityZone(config.CustomObject)
		i.Master.PrivateSubnet = key.SanitizeCFResourceName(key.PrivateSubnetName(i.Master.AZ))

		c := SmallCloudconfigConfig{
			S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&config.CustomObject, config.TenantClusterAccountID), key.S3ObjectPathTCCP(&config.CustomObject)),
		}
		rendered, err := template.Render(key.CloudConfigSmallTemplates(), c)
		if err != nil {
			return microerror.Mask(err)
		}
		i.Master.CloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

		i.Master.EncrypterBackend = config.EncrypterBackend
		i.Master.DockerVolume.Name = key.VolumeNameDocker(config.CustomObject)
		i.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
		i.Master.EtcdVolume.Name = key.VolumeNameEtcd(config.CustomObject)
		i.Master.LogVolume.Name = key.VolumeNameLog(config.CustomObject)
		i.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
		i.Master.Instance.Type = config.StackState.MasterInstanceType
		i.Master.Instance.Monitoring = config.StackState.MasterInstanceMonitoring
	}

	return nil
}

func (a *ParamsInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(&cfg.CustomObject)

	for _, az := range cfg.TenantClusterAvailabilityZones {
		ig := GuestInternetGatewayAdapterInternetGateway{
			InternetGatewayRoute: key.SanitizeCFResourceName(key.PublicInternetGatewayRouteName(az.Name)),
			RouteTable:           key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}

		a.InternetGateways = append(a.InternetGateways, ig)
	}

	return nil
}

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 2
	healthCheckInterval           = 5
	healthCheckTimeout            = 3
	healthCheckUnhealthyThreshold = 2
)

func (a *GuestLoadBalancersAdapter) Adapt(cfg Config) error {
	clusterAZs := cfg.TenantClusterAvailabilityZones
	if len(clusterAZs) < 1 {
		return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
	}

	// API load balancer settings.
	a.APIElbHealthCheckTarget = heathCheckTarget(key.KubernetesSecurePort)
	a.APIElbName = key.ELBNameAPI(&cfg.CustomObject)
	a.APIInternalElbName = key.InternalELBNameAPI(&cfg.CustomObject)
	a.APIElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      key.KubernetesSecurePort,
			PortInstance: key.KubernetesSecurePort,
		},
	}
	a.APIElbScheme = externalELBScheme
	a.APIInternalElbScheme = internalELBScheme

	// etcd load balancer settings.
	a.EtcdElbHealthCheckTarget = heathCheckTarget(key.EtcdPort)
	a.EtcdElbName = key.ELBNameEtcd(&cfg.CustomObject)
	a.EtcdElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      key.EtcdPort,
			PortInstance: key.EtcdPort,
		},
	}
	a.EtcdElbScheme = internalELBScheme

	// Ingress load balancer settings.
	a.IngressElbHealthCheckTarget = heathCheckTarget(key.IngressControllerSecurePort)
	a.IngressElbName = key.ELBNameIngress(&cfg.CustomObject)
	a.IngressElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB: httpsPort,

			PortInstance: key.IngressControllerSecurePort,
		},
		{
			PortELB:      httpPort,
			PortInstance: key.IngressControllerInsecurePort,
		},
	}
	a.IngressElbScheme = externalELBScheme

	// Load balancer health check settings.
	a.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	a.ELBHealthCheckInterval = healthCheckInterval
	a.ELBHealthCheckTimeout = healthCheckTimeout
	a.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold
	a.MasterInstanceResourceName = cfg.StackState.MasterInstanceResourceName

	for _, az := range clusterAZs {
		a.PublicSubnets = append(a.PublicSubnets, key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)))
	}

	for _, az := range clusterAZs {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		a.PrivateSubnets = append(a.PrivateSubnets, key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name)))
	}

	return nil
}

func heathCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}

func (a *GuestNATGatewayAdapter) Adapt(cfg Config) error {
	for _, az := range cfg.TenantClusterAvailabilityZones {
		gw := Gateway{
			AvailabilityZone: az.Name,
			NATGWName:        key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATEIPName:       key.SanitizeCFResourceName(key.NATEIPName(az.Name)),
			PublicSubnetName: key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		nr := NATRoute{
			NATGWName:             key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATRouteName:          key.SanitizeCFResourceName(key.NATRouteName(az.Name)),
			PrivateRouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		a.NATRoutes = append(a.NATRoutes, nr)
	}

	return nil
}

func (a *GuestOutputsAdapter) Adapt(config Config) error {
	a.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
	a.Master.ImageID = config.StackState.MasterImageID
	a.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
	a.Master.Instance.Type = config.StackState.MasterInstanceType

	a.OperatorVersion = config.StackState.OperatorVersion

	a.Route53Enabled = config.Route53Enabled

	return nil
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = key.ClusterBaseDomain(config.CustomObject)
	a.EtcdDomain = key.ClusterEtcdEndpoint(config.CustomObject)
	a.ClusterID = key.ClusterID(&config.CustomObject)
	a.MasterInstanceResourceName = config.StackState.MasterInstanceResourceName
	a.Route53Enabled = config.Route53Enabled
	a.VPCRegion = key.Region(config.CustomObject)

	return nil
}

const (
	allPorts             = -1
	cadvisorPort         = 4194
	etcdPort             = 2379
	kubeletPort          = 10250
	nodeExporterPort     = 10300
	kubeStateMetricsPort = 10301
	sshPort              = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"

	ingressSecurityGroupName = "IngressSecurityGroup"
)

func (s *GuestSecurityGroupsAdapter) Adapt(cfg Config) error {
	masterRules, err := s.getMasterRules(cfg, cfg.ControlPlaneVPCCidr)
	if err != nil {
		return microerror.Mask(err)
	}

	internalAPIRules, err := getKubernetesPrivateAPIRules(cfg, cfg.ControlPlaneVPCCidr)
	if err != nil {
		return microerror.Mask(err)
	}

	s.APIWhitelistEnabled = cfg.APIWhitelist.Public.Enabled
	s.PrivateAPIWhitelistEnabled = cfg.APIWhitelist.Private.Enabled

	s.MasterSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "master")
	s.MasterSecurityGroupRules = masterRules

	s.IngressSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "ingress")
	s.IngressSecurityGroupRules = s.getIngressRules(cfg.CustomObject)

	s.EtcdELBSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "etcd-elb")
	s.EtcdELBSecurityGroupRules = s.getEtcdRules(cfg.CustomObject, cfg.ControlPlaneVPCCidr)

	s.APIInternalELBSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "internal-api")
	s.APIInternalELBSecurityGroupRules = internalAPIRules

	return nil
}

func (s *GuestSecurityGroupsAdapter) getMasterRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// Allow traffic to the Kubernetes API server depending on the API
	// whitelisting rules.
	publicAPIRules, err := getKubernetesPublicAPIRules(cfg, hostClusterCIDR)
	if err != nil {
		return []securityGroupRule{}, microerror.Mask(err)
	}

	// Other security group rules for the master.
	otherRules := []securityGroupRule{
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

func (s *GuestSecurityGroupsAdapter) getIngressRules(customObject v1alpha1.Cluster) []securityGroupRule {
	return []securityGroupRule{
		{
			Description: "Allow all http traffic to the ingress load balancer.",
			Port:        httpPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  defaultCIDR,
		},
		{
			Description: "Allow all https traffic to the ingress load balancer.",
			Port:        httpsPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  defaultCIDR,
		},
	}
}

func (s *GuestSecurityGroupsAdapter) getEtcdRules(customObject v1alpha1.Cluster, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
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

func getKubernetesPrivateAPIRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// When public API whitelisting is enabled, add separate security group rule per each subnet.
	if cfg.APIWhitelist.Private.Enabled {
		// Allow control-plane CIDR and tenant cluster CIDR.
		rules := []securityGroupRule{
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
				subnetRule := securityGroupRule{
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
		allowAllRule := []securityGroupRule{
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

func getKubernetesPublicAPIRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// When API whitelisting is enabled, add separate security group rule per each subnet.
	if cfg.APIWhitelist.Public.Enabled {
		rules := []securityGroupRule{
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
				subnetRule := securityGroupRule{
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
			return []securityGroupRule{}, microerror.Mask(err)
		}

		for _, gatewayRule := range hostClusterNATGatewayRules {
			rules = append(rules, gatewayRule)
		}

		return rules, nil
	} else {
		// When API whitelisting is disabled, allow all traffic.
		allowAllRule := []securityGroupRule{
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

func getHostClusterNATGatewayRules(cfg Config) ([]securityGroupRule, error) {
	var gatewayRules []securityGroupRule

	for _, address := range cfg.ControlPlaneNATGatewayAddresses {
		gatewayRule := securityGroupRule{
			Description: "Allow traffic from gateways.",
			Port:        key.KubernetesSecurePort,
			Protocol:    tcpProtocol,
			SourceCIDR:  fmt.Sprintf("%s/32", *address.PublicIp),
		}

		gatewayRules = append(gatewayRules, gatewayRule)
	}

	return gatewayRules, nil
}

func (s *GuestSubnetsAdapter) Adapt(cfg Config) error {
	zones := cfg.TenantClusterAvailabilityZones

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	{
		numAZs := len(zones)
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	for _, az := range zones {
		snetName := key.SanitizeCFResourceName(key.PublicSubnetName(az.Name))
		snet := Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR.String(),
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PublicSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
		}
		s.PublicSubnets = append(s.PublicSubnets, snet)
	}

	for _, az := range zones {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		snetName := key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name))
		snet := Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR.String(),
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PrivateSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
		}
		s.PrivateSubnets = append(s.PrivateSubnets, snet)
	}

	return nil
}

func (v *GuestVPCAdapter) Adapt(cfg Config) error {
	v.CidrBlock = key.StatusClusterNetworkCIDR(cfg.CustomObject)
	v.ClusterID = key.ClusterID(&cfg.CustomObject)
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.ControlPlaneAccountID
	v.PeerVPCID = cfg.ControlPlaneVPCID
	v.Region = key.Region(cfg.CustomObject)
	v.RegionARN = key.RegionARN(cfg.AWSRegion)
	v.PeerRoleArn = cfg.ControlPlanePeerRoleARN

	for _, az := range cfg.TenantClusterAvailabilityZones {
		rtName := RouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		rtName := RouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	return nil
}

type GuestAdapter struct {
	IAMPolicies     GuestIAMPoliciesAdapter
	InternetGateway GuestInternetGatewayAdapter
	Instance        GuestInstanceAdapter
	LoadBalancers   GuestLoadBalancersAdapter
	NATGateway      GuestNATGatewayAdapter
	Outputs         GuestOutputsAdapter
	RecordSets      GuestRecordSetsAdapter
	RouteTables     GuestRouteTablesAdapter
	SecurityGroups  GuestSecurityGroupsAdapter
	Subnets         GuestSubnetsAdapter
	VPC             GuestVPCAdapter
}
