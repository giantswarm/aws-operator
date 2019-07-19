package unittest

import (
	"context"
	"net"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
)

func DefaultContext() context.Context {
	cc := controllercontext.Context{
		Status: controllercontext.ContextStatus{
			ControlPlane: controllercontext.ContextStatusControlPlane{
				AWSAccountID: "control-plane-account",
				NATGateway:   controllercontext.ContextStatusControlPlaneNATGateway{},
				RouteTable:   controllercontext.ContextStatusControlPlaneRouteTable{},
				PeerRole: controllercontext.ContextStatusControlPlanePeerRole{
					ARN: "imaginary-cp-peer-role-arn",
				},
				VPC: controllercontext.ContextStatusControlPlaneVPC{
					CIDR: "10.1.0.0/16",
				},
			},
			TenantCluster: controllercontext.ContextStatusTenantCluster{
				AWS: controllercontext.ContextStatusTenantClusterAWS{
					AccountID: "tenant-account",
				},
				Encryption:            controllercontext.ContextStatusTenantClusterEncryption{},
				HostedZoneNameServers: "1.1.1.1,8.8.8.8",
				MasterInstance:        controllercontext.ContextStatusTenantClusterMasterInstance{},
				TCCP: controllercontext.ContextStatusTenantClusterTCCP{
					ASG: controllercontext.ContextStatusTenantClusterTCCPASG{},
					AvailabilityZones: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
						{
							Name: "eu-central-1a",
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
									ID:   "validPrivateSubnetID-1a",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.32/27"),
									ID:   "validPublicSubnetID-1a",
								},
							},
						},
						{
							Name: "eu-central-1b",
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
									ID:   "validPrivateSubnetID-1b",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.96/27"),
									ID:   "validPublicSubnetID-1b",
								},
							},
						},
						{
							Name: "eu-central-1c",
							Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.128/27"),
									ID:   "validPrivateSubnetID-1c",
								},
								Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.164/27"),
									ID:   "validPublicSubnetID-1c",
								},
							},
						},
					},
					IsTransitioning:   false,
					MachineDeployment: DefaultMachineDeployment(),
					VPC: controllercontext.ContextStatusTenantClusterTCCPVPC{
						ID:                  "imagenary-vpc-id",
						PeeringConnectionID: "imagenary-peering-connection-id",
					},
				},
				VersionBundleVersion: "6.3.0",
				WorkerInstance: controllercontext.ContextStatusTenantClusterWorkerInstance{
					DockerVolumeSizeGB: "100",
					Image:              "ami-0eb0d9bb7ad1bd1e9",
					Type:               "m5.xlarge",
				},
			},
		},
		Spec: controllercontext.ContextSpec{
			TenantCluster: controllercontext.ContextSpecTenantCluster{
				TCCP: controllercontext.ContextSpecTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
						{
							Name: "eu-central-1a",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
									ID:   "validPrivateSubnetID-1a",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.32/27"),
									ID:   "validPublicSubnetID-1a",
								},
							},
						},
						{
							Name: "eu-central-1b",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.64/27"),
									ID:   "validPrivateSubnetID-1b",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.96/27"),
									ID:   "validPublicSubnetID-1b",
								},
							},
						},
						{
							Name: "eu-central-1c",
							Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.128/27"),
									ID:   "validPrivateSubnetID-1c",
								},
								Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
									CIDR: mustParseCIDR("10.100.3.164/27"),
									ID:   "validPublicSubnetID-1c",
								},
							},
						},
					},
				},
				TCNP: controllercontext.ContextSpecTenantClusterTCNP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone{
						{
							Name: "eu-central-1a",
							Subnet: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnet{
								Private: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate{
									CIDR: mustParseCIDR("10.100.3.0/27"),
								},
							},
						},
						{
							Name: "eu-central-1b",
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
