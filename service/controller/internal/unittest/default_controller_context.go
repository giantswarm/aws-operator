package unittest

import (
	"context"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func ChinaControllerContext() controllercontext.Context {
	return controllercontext.Context{
		Spec: controllercontext.ContextSpec{
			TenantCluster: controllercontext.ContextSpecTenantCluster{
				TCCP: controllercontext.ContextSpecTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
						{
							Name: "cn-north-1a",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
									ID:   "private-subnet-id-1a",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.32/27"),
									ID:   "public-subnet-id-1a",
								},
							},
						},
						{
							Name: "cn-north-1b",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
									ID:   "private-subnet-id-1b",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.96/27"),
									ID:   "public-subnet-id-1b",
								},
							},
						},
						{
							Name: "cn-north-1c",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.128/27"),
									ID:   "private-subnet-id-1c",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.164/27"),
									ID:   "public-subnet-id-1c",
								},
							},
						},
					},
				},
				TCNP: controllercontext.ContextSpecTenantClusterTCNP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone{
						{
							Name: "cn-north-1a",
							NATGateway: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway{
								ID: "nat-gateway-id-cn-north-1a",
							},
							Subnet: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
								},
							},
						},
						{
							Name: "cn-north-1c",
							NATGateway: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway{
								ID: "nat-gateway-id-cn-north-1c",
							},
							Subnet: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
								},
							},
						},
					},
					SecurityGroupIDs: []string{"sg-test1"},
				},
			},
		},
		Status: controllercontext.ContextStatus{
			ControlPlane: controllercontext.ContextStatusControlPlane{
				AWSAccountID: "control-plane-account",
				NATGateway:   controllercontext.ContextStatusControlPlaneNATGateway{},
				RouteTables: []*ec2.RouteTable{
					{
						RouteTableId: aws.String("gauss-private-1-id"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("gauss-private-1-name"),
							},
						},
					},
					{
						RouteTableId: aws.String("gauss-private-2-id"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("gauss-private-2-name"),
							},
						},
					},
				},
				PeerRole: controllercontext.ContextStatusControlPlanePeerRole{
					ARN: "peer-role-arn",
				},
				VPC: controllercontext.ContextStatusControlPlaneVPC{
					CIDR: "10.1.0.0/16",
					ID:   "vpc-testid",
				},
			},
			TenantCluster: controllercontext.ContextStatusTenantCluster{
				AWS: controllercontext.ContextStatusTenantClusterAWS{
					AccountID: "tenant-account",
					Region:    "cn-north-1",
				},
				Encryption:            controllercontext.ContextStatusTenantClusterEncryption{},
				HostedZoneNameServers: "1.1.1.1,8.8.8.8",
				MasterInstance:        controllercontext.ContextStatusTenantClusterMasterInstance{},
				OperatorVersion:       "6.3.0",
				TCCP: controllercontext.ContextStatusTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
						{
							Name: "cn-north-1a",
							RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1a",
								},
							},
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
									ID:   "private-subnet-id-1a",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.32/27"),
									ID:   "public-subnet-id-1a",
								},
							},
						},
						{
							Name: "cn-north-1b",
							RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1b",
								},
							},
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
									ID:   "private-subnet-id-1b",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.96/27"),
									ID:   "public-subnet-id-1b",
								},
							},
						},
						{
							Name: "cn-north-1c",
							RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1c",
								},
							},
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.128/27"),
									ID:   "private-subnet-id-1c",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.164/27"),
									ID:   "public-subnet-id-1c",
								},
							},
						},
					},
					IsTransitioning: false,
					SecurityGroups: []*ec2.SecurityGroup{
						{
							GroupId: aws.String("ingress-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-ingress"),
								},
							},
						},
						{
							GroupId: aws.String("internal-api-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-internal-api"),
								},
							},
						},
						{
							GroupId: aws.String("master-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-master"),
								},
							},
						},
						{
							GroupId: aws.String("awscni-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-aws-cni"),
								},
							},
						},
					},
					Subnets: []*ec2.Subnet{
						{
							SubnetId: aws.String("subnet-id-eu-central-1a"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("PrivateSubnetEuCentral1a"),
								},
							},
						},
						{
							SubnetId: aws.String("subnet-id-eu-central-1b"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("PrivateSubnetEuCentral1b"),
								},
							},
						},
						{
							SubnetId: aws.String("subnet-id-eu-central-1c"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("PrivateSubnetEuCentral1c"),
								},
							},
						},
					},
					VPC: controllercontext.ContextStatusTenantClusterTCCPVPC{
						ID:                  "vpc-id",
						PeeringConnectionID: "peering-connection-id",
					},
				},
			},
		},
	}
}

func DefaultControllerContext() controllercontext.Context {
	return controllercontext.Context{
		Spec: controllercontext.ContextSpec{
			TenantCluster: controllercontext.ContextSpecTenantCluster{
				TCCP: controllercontext.ContextSpecTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
						{
							Name: "eu-central-1a",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
									ID:   "private-subnet-id-1a",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.32/27"),
									ID:   "public-subnet-id-1a",
								},
							},
						},
						{
							Name: "eu-central-1b",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
									ID:   "private-subnet-id-1b",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.96/27"),
									ID:   "public-subnet-id-1b",
								},
							},
						},
						{
							Name: "eu-central-1c",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.128/27"),
									ID:   "private-subnet-id-1c",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.164/27"),
									ID:   "public-subnet-id-1c",
								},
							},
						},
					},
				},
				TCNP: controllercontext.ContextSpecTenantClusterTCNP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone{
						{
							Name: "eu-central-1a",
							NATGateway: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway{
								ID: "nat-gateway-id-eu-central-1a",
							},
							Subnet: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
								},
							},
						},
						{
							Name: "eu-central-1c",
							NATGateway: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway{
								ID: "nat-gateway-id-eu-central-1c",
							},
							Subnet: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
								},
							},
						},
					},
					SecurityGroupIDs: []string{"sg-test1"},
				},
			},
		},
		Status: controllercontext.ContextStatus{
			ControlPlane: controllercontext.ContextStatusControlPlane{
				AWSAccountID: "control-plane-account",
				NATGateway:   controllercontext.ContextStatusControlPlaneNATGateway{},
				RouteTables: []*ec2.RouteTable{
					{
						RouteTableId: aws.String("gauss-private-1-id"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("gauss-private-1-name"),
							},
						},
					},
					{
						RouteTableId: aws.String("gauss-private-2-id"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("gauss-private-2-name"),
							},
						},
					},
				},
				PeerRole: controllercontext.ContextStatusControlPlanePeerRole{
					ARN: "peer-role-arn",
				},
				VPC: controllercontext.ContextStatusControlPlaneVPC{
					CIDR: "10.1.0.0/16",
					ID:   "vpc-testid",
				},
			},
			TenantCluster: controllercontext.ContextStatusTenantCluster{
				AWS: controllercontext.ContextStatusTenantClusterAWS{
					AccountID: "tenant-account",
					Region:    "eu-central-1",
				},
				Encryption:            controllercontext.ContextStatusTenantClusterEncryption{},
				HostedZoneNameServers: "1.1.1.1,8.8.8.8",
				MasterInstance:        controllercontext.ContextStatusTenantClusterMasterInstance{},
				OperatorVersion:       "6.3.0",
				TCCP: controllercontext.ContextStatusTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
						{
							Name: "eu-central-1a",
							RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1a",
								},
							},
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
									ID:   "private-subnet-id-1a",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.32/27"),
									ID:   "public-subnet-id-1a",
								},
							},
						},
						{
							Name: "eu-central-1b",
							RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1b",
								},
							},
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
									ID:   "private-subnet-id-1b",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.96/27"),
									ID:   "public-subnet-id-1b",
								},
							},
						},
						{
							Name: "eu-central-1c",
							RouteTable: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1c",
								},
							},
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.128/27"),
									ID:   "private-subnet-id-1c",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.164/27"),
									ID:   "public-subnet-id-1c",
								},
							},
						},
					},
					IsTransitioning: false,
					SecurityGroups: []*ec2.SecurityGroup{
						{
							GroupId: aws.String("ingress-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-ingress"),
								},
							},
						},
						{
							GroupId: aws.String("internal-api-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-internal-api"),
								},
							},
						},
						{
							GroupId: aws.String("master-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-master"),
								},
							},
						},
						{
							GroupId: aws.String("awscni-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-aws-cni"),
								},
							},
						},
					},
					Subnets: []*ec2.Subnet{
						{
							SubnetId: aws.String("subnet-id-eu-central-1a"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("PrivateSubnetEuCentral1a"),
								},
							},
						},
						{
							SubnetId: aws.String("subnet-id-eu-central-1b"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("PrivateSubnetEuCentral1b"),
								},
							},
						},
						{
							SubnetId: aws.String("subnet-id-eu-central-1c"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("PrivateSubnetEuCentral1c"),
								},
							},
						},
					},
					VPC: controllercontext.ContextStatusTenantClusterTCCPVPC{
						ID:                  "vpc-id",
						PeeringConnectionID: "peering-connection-id",
					},
				},
			},
		},
	}
}

func ChinaContext() context.Context {
	cc := ChinaControllerContext()
	return controllercontext.NewContext(context.Background(), cc)
}

func DefaultContext() context.Context {
	cc := DefaultControllerContext()
	return controllercontext.NewContext(context.Background(), cc)
}

func mustParseCIDR(s string) net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *n
}
