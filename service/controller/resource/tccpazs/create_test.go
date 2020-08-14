package tccpazs

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/to"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/unittest"
)

func Test_EnsureCreated_AZ_Spec(t *testing.T) {
	testCases := []struct {
		name               string
		cluster            infrastructurev1alpha2.AWSCluster
		controlPlane       infrastructurev1alpha2.AWSControlPlane
		machineDeployments []infrastructurev1alpha2.AWSMachineDeployment
		ctxStatusSubnets   []*ec2.Subnet
		expectedAZs        []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone
		errorMatcher       func(error) bool
	}{
		{
			name:               "case 0: keep control plane, 0 node pools",
			cluster:            unittest.ClusterWithNetworkCIDR(unittest.DefaultCluster(), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			controlPlane:       unittest.DefaultAWSControlPlaneWithAZs("eu-central-1a"),
			machineDeployments: []infrastructurev1alpha2.AWSMachineDeployment{},
			ctxStatusSubnets: []*ec2.Subnet{
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.0/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.32/27"),
				},
			},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.100.3.32/27"),
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.100.3.0/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 1: control plane and 1 node pool on same AZ",
			cluster:      unittest.ClusterWithNetworkCIDR(unittest.DefaultCluster(), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			controlPlane: unittest.DefaultAWSControlPlaneWithAZs("eu-central-1a"),
			machineDeployments: []infrastructurev1alpha2.AWSMachineDeployment{
				unittest.MachineDeploymentWithAZs(unittest.DefaultMachineDeployment(), []string{"eu-central-1a"}),
			},
			ctxStatusSubnets: []*ec2.Subnet{
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.0/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.32/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.5.0/24"),
				},
			},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.100.3.32/27"),
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.100.3.0/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 2: create control plane and 1 node pool on different AZ",
			cluster:      unittest.ClusterWithNetworkCIDR(unittest.DefaultCluster(), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			controlPlane: unittest.DefaultAWSControlPlaneWithAZs("eu-central-1a"),
			machineDeployments: []infrastructurev1alpha2.AWSMachineDeployment{
				unittest.MachineDeploymentWithAZs(unittest.DefaultMachineDeployment(), []string{"eu-central-1b"}),
			},
			ctxStatusSubnets: []*ec2.Subnet{},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.100.3.32/27"),
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.100.3.0/27"),
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.100.3.96/27"),
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.100.3.64/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:               "case 3: keep control plane and delete 1 node pool from different AZ",
			cluster:            unittest.ClusterWithNetworkCIDR(unittest.DefaultCluster(), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			controlPlane:       unittest.DefaultAWSControlPlaneWithAZs("eu-central-1a"),
			machineDeployments: []infrastructurev1alpha2.AWSMachineDeployment{},
			ctxStatusSubnets: []*ec2.Subnet{
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.0/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.32/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1b"),
					CidrBlock:        aws.String("10.100.3.64/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1b"),
					CidrBlock:        aws.String("10.100.3.96/27"),
				},
				{
					AvailabilityZone: aws.String("eu-central-1b"),
					CidrBlock:        aws.String("10.100.5.0/24"),
				},
			},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.100.3.32/27"),
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.100.3.0/27"),
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
						AWSCNI: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
						Private: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate{
							CIDR: mustParseCIDR("10.100.3.96/27"),
						},
						Public: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic{
							CIDR: mustParseCIDR("10.100.3.64/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContext()

			var r *Resource
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),
					Logger:    microloggertest.New(),

					CIDRBlockAWSCNI: "172.17.0.0/16",
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Prepare all the necessary runtime objects using the abstract controller
			// client.
			{
				err = r.k8sClient.CtrlClient().Create(ctx, &tc.cluster)
				if err != nil {
					t.Fatal(err)
				}

				err = r.k8sClient.CtrlClient().Create(ctx, &tc.controlPlane)
				if err != nil {
					t.Fatal(err)
				}

				for _, md := range tc.machineDeployments {
					err := r.k8sClient.CtrlClient().Create(ctx, &md)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			cc, err := controllercontext.FromContext(ctx)
			if err != nil {
				t.Fatal(err)
			}
			cc.Status.TenantCluster.TCCP.Subnets = tc.ctxStatusSubnets

			err = r.EnsureCreated(ctx, &tc.cluster)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			cc, err = controllercontext.FromContext(ctx)
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.expectedAZs, cc.Spec.TenantCluster.TCCP.AvailabilityZones)
			if diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
		})
	}
}

func Test_ensureAZsAreAssignedWithSubnet(t *testing.T) {
	testCases := []struct {
		name         string
		awsCNISubnet net.IPNet
		tccpSubnet   net.IPNet
		inputAZs     map[string]mapping
		expectedAZs  map[string]mapping
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: three AZs without subnets",
			awsCNISubnet: mustParseCIDR("172.17.0.0/16"),
			tccpSubnet:   mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {},
				"eu-central-1b": {},
				"eu-central-1c": {},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.96/27"),
						},
					},
				},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.128.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 1: three AZs, one without subnets",
			awsCNISubnet: mustParseCIDR("172.17.0.0/16"),
			tccpSubnet:   mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.128.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.96/27"),
						},
					},
				},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 2: three AZs, two without subnets",
			awsCNISubnet: mustParseCIDR("172.17.0.0/16"),
			tccpSubnet:   mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {},
				"eu-central-1b": {},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.128.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.96/27"),
						},
					},
				},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 3: four AZs, one without subnets",
			awsCNISubnet: mustParseCIDR("172.17.0.0/16"),
			tccpSubnet:   mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.96/27"),
						},
					},
				},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.128.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
				"eu-central-1d": {},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.96/27"),
						},
					},
				},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.128.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
				"eu-central-1d": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.192.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.192/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.224/27"),
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:         "case 4: five AZs, one without subnets",
			awsCNISubnet: mustParseCIDR("172.17.0.0/16"),
			tccpSubnet:   mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.0.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
						},
					},
				},
				"eu-central-1b": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.64.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.96/27"),
						},
					},
				},
				"eu-central-1c": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.128.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.128/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.160/27"),
						},
					},
				},
				"eu-central-1d": {
					AWSCNI: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("172.17.192.0/18"),
						},
					},
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.192/27"),
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.224/27"),
						},
					},
				},
				"eu-central-1e": {},
			},
			expectedAZs:  nil,
			errorMatcher: IsInvalidConfig,
		},
	}

	var r *Resource
	{
		var err error

		c := Config{
			K8sClient: unittest.FakeK8sClient(),
			Logger:    microloggertest.New(),

			CIDRBlockAWSCNI: "dummy",
		}

		r, err = New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := controllercontext.NewContext(context.Background(), controllercontext.Context{})
			azs, err := r.ensureAZsAreAssignedWithSubnet(ctx, tc.awsCNISubnet, tc.tccpSubnet, tc.inputAZs)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			diff := cmp.Diff(azs, tc.expectedAZs)

			if diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
		})
	}
}

func Test_mapSubnets(t *testing.T) {
	testCases := []struct {
		name         string
		input        []*ec2.Subnet
		expected     map[string]mapping
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: empty list of subnets",
			input:        nil,
			expected:     make(map[string]mapping),
			errorMatcher: nil,
		},
		{
			name: "case 1: network list with one value",
			input: []*ec2.Subnet{
				{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]mapping{
				"eu-central-1a": {
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
							ID:   "validID",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: subnet list with one value",
			input: []*ec2.Subnet{
				{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]mapping{
				"eu-central-1a": {
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
							ID:   "validID",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: subnet list with three values",
			input: []*ec2.Subnet{
				{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
				{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.32/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("private"),
						},
					},
				},
				{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        to.StringP("10.100.8.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]mapping{
				"eu-central-1a": {
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
							ID:   "validID",
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
							ID:   "validID",
						},
					},
				},
				"eu-central-1b": {
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.64/27"),
							ID:   "validID",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 4: subnet list with irrelevant values",
			input: []*ec2.Subnet{
				{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
				{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.32/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						{
							Key:   to.StringP(key.TagStack),
							Value: to.StringP(key.StackTCCP),
						},
						{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("private"),
						},
					},
				},
				{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        to.StringP("10.100.4.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags:             []*ec2.Tag{},
				},
			},
			expected: map[string]mapping{
				"eu-central-1a": {
					Public: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.0/27"),
							ID:   "validID",
						},
					},
					Private: network{
						Subnet: subnet{
							CIDR: mustParseCIDR("10.100.8.32/27"),
							ID:   "validID",
						},
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output, err := mapSubnets(map[string]mapping{}, tc.input)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			diff := cmp.Diff(output, tc.expected)

			if diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
		})
	}
}

func mustParseCIDR(v string) net.IPNet {
	_, n, err := net.ParseCIDR(v)
	if err != nil {
		panic(err)
	}
	return *n
}

func toNetPtr(n net.IPNet) *net.IPNet {
	return &n
}
