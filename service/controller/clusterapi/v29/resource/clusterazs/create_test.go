package clusterazs

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/to"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func Test_ensureAZsAreAssignedWithSubnet(t *testing.T) {
	testCases := []struct {
		name         string
		tccpSubnet   net.IPNet
		inputAZs     map[string]subnetPair
		expectedAZs  map[string]subnetPair
		errorMatcher func(error) bool
	}{
		{
			name:       "case 0: three AZs without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{},
				"eu-central-1b": subnetPair{},
				"eu-central-1c": subnetPair{},
			},
			expectedAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.96/27"),
					},
				},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 1: three AZs, one without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
			},
			expectedAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.96/27"),
					},
				},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 2: three AZs, two without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{},
				"eu-central-1b": subnetPair{},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
			},
			expectedAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.96/27"),
					},
				},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 3: four AZs, one without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.96/27"),
					},
				},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
				"eu-central-1d": subnetPair{},
			},
			expectedAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.96/27"),
					},
				},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
				"eu-central-1d": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.192/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.224/27"),
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:       "case 4: five AZs, one without subnets",
			tccpSubnet: mustParseCIDR("10.100.8.0/24"),
			inputAZs: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.96/27"),
					},
				},
				"eu-central-1c": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.128/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.160/27"),
					},
				},
				"eu-central-1d": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.192/27"),
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.224/27"),
					},
				},
				"eu-central-1e": subnetPair{},
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

func Test_fromEC2SubnetsToMap(t *testing.T) {
	testCases := []struct {
		name         string
		input        []*ec2.Subnet
		expected     map[string]subnetPair
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: empty list of subnets",
			input:        nil,
			expected:     make(map[string]subnetPair),
			errorMatcher: nil,
		},
		{
			name: "case 1: subnet list with one value",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
						ID:   "validID",
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: subnet list with one value",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
						ID:   "validID",
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: subnet list with three values",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.32/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("private"),
						},
					},
				},
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        to.StringP("10.100.8.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
						ID:   "validID",
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
						ID:   "validID",
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
						ID:   "validID",
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: subnet list with irrelevant values",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.32/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("private"),
						},
					},
				},
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        to.StringP("10.100.4.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags:             []*ec2.Tag{},
				},
			},
			expected: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
						ID:   "validID",
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
						ID:   "validID",
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 4: subnet list with subnet that is nil",
			input: []*ec2.Subnet{
				nil,
			},
			expected:     nil,
			errorMatcher: isExecutionFailed,
		},
		{
			name: "case 5: subnet list with subnet that has AvailabilityZone nil",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: nil,
					CidrBlock:        to.StringP("10.100.4.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected:     nil,
			errorMatcher: isExecutionFailed,
		},
		{
			name: "case 6: subnet list with subnet that has CidrBlock nil",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        nil,
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected:     nil,
			errorMatcher: isExecutionFailed,
		},
		{
			name: "case 7: subnet list with subnet that has tags nil",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        to.StringP("10.100.4.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags:             nil,
				},
			},
			expected:     nil,
			errorMatcher: isExecutionFailed,
		},
		{
			name: "case 8: subnet list with subnet that has Tag that is nil",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        nil,
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						nil,
					},
				},
			},
			expected:     nil,
			errorMatcher: isExecutionFailed,
		},
		{
			name: "case 9: subnet list with subnet that has Tag that has nil Key",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        nil,
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   nil,
							Value: to.StringP("orphaned value"),
						},
					},
				},
			},
			expected:     nil,
			errorMatcher: isExecutionFailed,
		},
		{
			name: "case 10: subnet list with three values and one that has Tag that has nil Value",
			input: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.0/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP("uninteresting-tag"),
							Value: nil,
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1a"),
					CidrBlock:        to.StringP("10.100.8.32/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("private"),
						},
					},
				},
				&ec2.Subnet{
					AvailabilityZone: to.StringP("eu-central-1b"),
					CidrBlock:        to.StringP("10.100.8.64/27"),
					SubnetId:         to.StringP("validID"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   to.StringP(key.TagTCCP),
							Value: to.StringP("true"),
						},
						&ec2.Tag{
							Key:   to.StringP(key.TagSubnetType),
							Value: to.StringP("public"),
						},
					},
				},
			},
			expected: map[string]subnetPair{
				"eu-central-1a": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.0/27"),
						ID:   "validID",
					},
					Private: subnet{
						CIDR: mustParseCIDR("10.100.8.32/27"),
						ID:   "validID",
					},
				},
				"eu-central-1b": subnetPair{
					Public: subnet{
						CIDR: mustParseCIDR("10.100.8.64/27"),
						ID:   "validID",
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output, err := fromEC2SubnetsToMap(tc.input)

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
