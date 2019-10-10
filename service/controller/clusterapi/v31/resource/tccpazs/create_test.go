package tccpazs

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/to"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/unittest"
)

func Test_EnsureCreated_AZ_Spec(t *testing.T) {
	testCases := []struct {
		name               string
		cluster            cmav1alpha1.Cluster
		machineDeployments []cmav1alpha1.MachineDeployment
		ctxStatusSubnets   []*ec2.Subnet
		expectedAZs        []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone
		errorMatcher       func(error) bool
	}{
		{
			name:               "case 0: keep control plane, 0 node pools",
			cluster:            unittest.ClusterWithNetworkCIDR(unittest.ClusterWithAZ(unittest.DefaultCluster(), "eu-central-1a"), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			machineDeployments: nil,
			ctxStatusSubnets: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.0/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.32/27"),
				},
			},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
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
			name:    "case 1: control plane and 1 node pool on same AZ",
			cluster: unittest.ClusterWithNetworkCIDR(unittest.ClusterWithAZ(unittest.DefaultCluster(), "eu-central-1a"), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			machineDeployments: []cmav1alpha1.MachineDeployment{
				unittest.MachineDeploymentWithAZs(unittest.DefaultMachineDeployment(), []string{"eu-central-1a"}),
			},
			ctxStatusSubnets: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.0/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.32/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.5.0/24"),
				},
			},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
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
			name:    "case 2: create control plane and 1 node pool on different AZ",
			cluster: unittest.ClusterWithNetworkCIDR(unittest.ClusterWithAZ(unittest.DefaultCluster(), "eu-central-1a"), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			machineDeployments: []cmav1alpha1.MachineDeployment{
				unittest.MachineDeploymentWithAZs(unittest.DefaultMachineDeployment(), []string{"eu-central-1b"}),
			},
			ctxStatusSubnets: []*ec2.Subnet{},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
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
			cluster:            unittest.ClusterWithNetworkCIDR(unittest.ClusterWithAZ(unittest.DefaultCluster(), "eu-central-1a"), toNetPtr(mustParseCIDR("10.100.3.0/24"))),
			machineDeployments: []cmav1alpha1.MachineDeployment{},
			ctxStatusSubnets: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.0/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1a"),
					CidrBlock:        aws.String("10.100.3.32/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1b"),
					CidrBlock:        aws.String("10.100.3.64/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1b"),
					CidrBlock:        aws.String("10.100.3.96/27"),
				},
				&ec2.Subnet{
					AvailabilityZone: aws.String("eu-central-1b"),
					CidrBlock:        aws.String("10.100.5.0/24"),
				},
			},
			expectedAZs: []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: controllercontext.ContextSpecTenantClusterTCCPAvailabilityZoneSubnet{
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
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Construct fresh fake client for each test case.
			fakeClient := fake.NewSimpleClientset()

			var r *Resource
			{
				var err error

				c := Config{
					CMAClient:     fakeClient,
					Logger:        microloggertest.New(),
					ToClusterFunc: key.ToCluster,
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Prepare MachineDeployments for fake client.
			for _, md := range tc.machineDeployments {
				_, err := fakeClient.ClusterV1alpha1().MachineDeployments(metav1.NamespaceDefault).Create(&md)
				if err != nil {
					t.Fatal(err)
				}
			}

			ctx := unittest.DefaultContext()
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

			diff := cmp.Diff(cc.Spec.TenantCluster.TCCP.AvailabilityZones, tc.expectedAZs)

			if diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
		})
	}

}

func Test_ensureAZsAreAssignedWithSubnet(t *testing.T) {
	testCases := []struct {
		name         string
		tccpSubnet   net.IPNet
		inputAZs     map[string]mapping
		expectedAZs  map[string]mapping
		errorMatcher func(error) bool
	}{
		{
			name:       "case 0: three AZs without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
					RequiredByCR: true,
				},
				"eu-central-1b": {
					RequiredByCR: true,
				},
				"eu-central-1c": {
					RequiredByCR: true,
				},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
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
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 1: three AZs, one without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
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
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 2: three AZs, two without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
					RequiredByCR: true,
				},
				"eu-central-1b": {
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
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
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 3: four AZs, one without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
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
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
				"eu-central-1d": {
					RequiredByCR: true,
				},
			},
			expectedAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
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
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
				"eu-central-1d": {
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
					RequiredByCR: true,
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 4: five AZs, one without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]mapping{
				"eu-central-1a": {
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
					RequiredByCR: true,
				},
				"eu-central-1b": {
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
					RequiredByCR: true,
				},
				"eu-central-1c": {
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
					RequiredByCR: true,
				},
				"eu-central-1d": {
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
					RequiredByCR: true,
				},
				"eu-central-1e": {
					RequiredByCR: true,
				},
			},
			expectedAZs:  nil,
			errorMatcher: IsInvalidConfig,
		},
	}

	var r *Resource
	{
		var err error

		c := Config{
			CMAClient:     fake.NewSimpleClientset(),
			Logger:        microloggertest.New(),
			ToClusterFunc: key.ToCluster,
		}

		r, err = New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := controllercontext.NewContext(context.Background(), controllercontext.Context{})
			azs, err := r.ensureAZsAreAssignedWithSubnet(ctx, tc.tccpSubnet, tc.inputAZs)

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

func isExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
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
