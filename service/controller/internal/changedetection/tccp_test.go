package changedetection

import (
	"net"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func Test_Detection_availabilityZonesEqual(t *testing.T) {
	testCases := []struct {
		name    string
		spec    []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone
		status  []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone
		matches bool
	}{
		{
			name:    "case 0",
			spec:    nil,
			status:  nil,
			matches: true,
		},
		{
			name:    "case 1",
			spec:    []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{},
			status:  []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{},
			matches: true,
		},
		{
			name:    "case 2",
			spec:    nil,
			status:  []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{},
			matches: false,
		},
		{
			name:    "case 3",
			spec:    []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{},
			status:  nil,
			matches: false,
		},
		{
			name: "case 4",
			spec: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
			},
			status: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
			},
			matches: true,
		},
		{
			name: "case 5",
			spec: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.10.40.0/27"), // DIFFERENT
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
			},
			status: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
			},
			matches: false,
		},
		{
			name: "case 6",
			spec: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.64/27"),
							ID:   "subnet-0934681f126016726",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.96/27"),
							ID:   "subnet-0debe88c7b8120cb4",
						},
					},
				},
			},
			status: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.64/27"),
							ID:   "subnet-0934681f126016726",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.96/27"),
							ID:   "subnet-0debe88c7b8120cb4",
						},
					},
				},
			},
			matches: true,
		},
		{
			name: "case 7",
			spec: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.64/27"),
							ID:   "subnet-0934681f126016726",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.96/27"),
							ID:   "subnet-0debe88c7b8120cb4",
						},
					},
				},
			},
			status: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-3557f0i4c2ne3ef88", // DIFFERENT
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.64/27"),
							ID:   "subnet-0934681f126016726",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.96/27"),
							ID:   "subnet-0debe88c7b8120cb4",
						},
					},
				},
			},
			matches: false,
		},
		{
			name: "case 8, different order",
			spec: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.64/27"),
							ID:   "subnet-0934681f126016726",
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.96/27"),
							ID:   "subnet-0debe88c7b8120cb4",
						},
					},
				},
			},
			status: []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.64/27"),
							ID:   "subnet-0934681f126016726",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.96/27"),
							ID:   "subnet-0debe88c7b8120cb4",
						},
					},
				},
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnet{
						Private: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.1.4.0/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
						Public: controllercontext.ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.1.4.32/27"),
							ID:   "subnet-0854f0e4c66e3ef10",
						},
					},
				},
			},
			matches: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			matches := availabilityZonesEqual(tc.spec, tc.status)

			if matches != tc.matches {
				t.Fatalf("\n\n%s\n", cmp.Diff(matches, tc.matches))
			}
		})
	}
}

func mustParseCIDR(s string) net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *n
}
