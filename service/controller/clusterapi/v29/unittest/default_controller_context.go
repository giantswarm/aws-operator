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
				AWSAccountID:          "tenant-account",
				Encryption:            controllercontext.ContextStatusTenantClusterEncryption{},
				HostedZoneNameServers: "1.1.1.1,8.8.8.8",
				MasterInstance:        controllercontext.ContextStatusTenantClusterMasterInstance{},
				TCCP: controllercontext.ContextStatusTenantClusterTCCP{
					ASG: controllercontext.ContextStatusTenantClusterTCCPASG{},
					AvailabilityZones: []controllercontext.ContextTenantClusterAvailabilityZone{
						{
							Name: "eu-central-1a",
							PrivateSubnet: controllercontext.ContextTenantClusterAvailabilityZonePrivateSubnet{
								CIDR: mustParseCIDR("10.100.3.0/27"),
								ID:   "validPrivateSubnetID-1a",
							},
							PublicSubnet: controllercontext.ContextTenantClusterAvailabilityZonePublicSubnet{
								CIDR: mustParseCIDR("10.100.3.32/27"),
								ID:   "validPublicSubnetID-1a",
							},
						},
						{
							Name: "eu-central-1b",
							PrivateSubnet: controllercontext.ContextTenantClusterAvailabilityZonePrivateSubnet{
								CIDR: mustParseCIDR("10.100.3.64/27"),
								ID:   "validPrivateSubnetID-1b",
							},
							PublicSubnet: controllercontext.ContextTenantClusterAvailabilityZonePublicSubnet{
								CIDR: mustParseCIDR("10.100.3.96/27"),
								ID:   "validPublicSubnetID-1b",
							},
						},
						{
							Name: "eu-central-1c",
							PrivateSubnet: controllercontext.ContextTenantClusterAvailabilityZonePrivateSubnet{
								CIDR: mustParseCIDR("10.100.3.128/27"),
								ID:   "validPrivateSubnetID-1c",
							},
							PublicSubnet: controllercontext.ContextTenantClusterAvailabilityZonePublicSubnet{
								CIDR: mustParseCIDR("10.100.3.164/27"),
								ID:   "validPublicSubnetID-1c",
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
					AvailabilityZones: []controllercontext.ContextTenantClusterAvailabilityZone{
						{
							Name: "eu-central-1a",
							PrivateSubnet: controllercontext.ContextTenantClusterAvailabilityZonePrivateSubnet{
								CIDR: mustParseCIDR("10.100.3.0/27"),
								ID:   "validPrivateSubnetID-1a",
							},
							PublicSubnet: controllercontext.ContextTenantClusterAvailabilityZonePublicSubnet{
								CIDR: mustParseCIDR("10.100.3.32/27"),
								ID:   "validPublicSubnetID-1a",
							},
						},
						{
							Name: "eu-central-1b",
							PrivateSubnet: controllercontext.ContextTenantClusterAvailabilityZonePrivateSubnet{
								CIDR: mustParseCIDR("10.100.3.64/27"),
								ID:   "validPrivateSubnetID-1b",
							},
							PublicSubnet: controllercontext.ContextTenantClusterAvailabilityZonePublicSubnet{
								CIDR: mustParseCIDR("10.100.3.96/27"),
								ID:   "validPublicSubnetID-1b",
							},
						},
						{
							Name: "eu-central-1c",
							PrivateSubnet: controllercontext.ContextTenantClusterAvailabilityZonePrivateSubnet{
								CIDR: mustParseCIDR("10.100.3.128/27"),
								ID:   "validPrivateSubnetID-1c",
							},
							PublicSubnet: controllercontext.ContextTenantClusterAvailabilityZonePublicSubnet{
								CIDR: mustParseCIDR("10.100.3.164/27"),
								ID:   "validPublicSubnetID-1c",
							},
						},
					},
				},
				TCNP: controllercontext.ContextSpecTenantClusterTCNP{
					AvailabilityZones: []controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone{
						{
							AvailabilityZone: "eu-central-1a",
							PrivateSubnet:    mustParseCIDR("10.100.3.0/27"),
						},
						{
							AvailabilityZone: "eu-central-1b",
							PrivateSubnet:    mustParseCIDR("10.100.3.64/27"),
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
