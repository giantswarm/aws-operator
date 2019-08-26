package unittest

import (
	"context"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
)

func DefaultContext() context.Context {
	cc := controllercontext.Context{
		Spec: controllercontext.ContextSpec{
			TenantCluster: controllercontext.ContextSpecTenantCluster{
				TCCP: controllercontext.ContextSpecTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
						{
							Name: "eu-central-1a",
							RouteTable: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1a",
								},
							},
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
							RouteTable: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1b",
								},
							},
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
							RouteTable: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTable{
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic{
									ID: "public-route-table-id-1c",
								},
							},
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
					IsTransitioning:   false,
					MachineDeployment: DefaultMachineDeployment(),
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
							GroupId: aws.String("master-security-group-id"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("8y5ck-master"),
								},
							},
						},
					},
					VPC: controllercontext.ContextStatusTenantClusterTCCPVPC{
						ID:                  "vpc-id",
						PeeringConnectionID: "peering-connection-id",
					},
				},
				TCNP: controllercontext.ContextStatusTenantClusterTCNP{
					ASG: controllercontext.ContextStatusTenantClusterTCNPASG{},
				},
				VersionBundleVersion: "6.3.0",
			},
		},
	}

	return controllercontext.NewContext(context.Background(), cc)
}

func mustParseCIDR(s string) net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *n
}
