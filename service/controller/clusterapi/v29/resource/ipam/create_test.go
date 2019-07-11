package ipam

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func Test_SubnetAllocator(t *testing.T) {
	testCases := []struct {
		name         string
		object       interface{}
		errorMatcher func(error) bool

		checker   Checker
		collector Collector
		persister Persister

		allocatedSubnetMaskBits int
		networkRange            net.IPNet
		privateSubnetMaskBits   int
		publicSubnetMaskBits    int
	}{
		{
			name:         "case 0 allocate first subnet",
			object:       &v1alpha1.Cluster{},
			errorMatcher: nil,

			checker:   NewTestChecker(true),
			collector: NewTestCollector([]net.IPNet{}),
			persister: NewTestPersister(mustParseCIDR("10.100.0.0/24")),

			allocatedSubnetMaskBits: 24,
			networkRange:            mustParseCIDR("10.100.0.0/16"),
			privateSubnetMaskBits:   25,
			publicSubnetMaskBits:    25,
		},
		{
			name:         "case 1 allocate fourth subnet",
			object:       &v1alpha1.Cluster{},
			errorMatcher: nil,

			checker: NewTestChecker(true),
			collector: NewTestCollector([]net.IPNet{
				mustParseCIDR("10.100.0.0/24"),
				mustParseCIDR("10.100.1.0/24"),
				mustParseCIDR("10.100.3.0/24"),
			}),
			persister: NewTestPersister(mustParseCIDR("10.100.2.0/24")),

			allocatedSubnetMaskBits: 24,
			networkRange:            mustParseCIDR("10.100.0.0/16"),
			privateSubnetMaskBits:   25,
			publicSubnetMaskBits:    25,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var newResource *Resource
			{
				c := Config{
					Checker:   tc.checker,
					Collector: tc.collector,
					Logger:    microloggertest.New(),
					Persister: tc.persister,

					AllocatedSubnetMaskBits: tc.allocatedSubnetMaskBits,
					NetworkRange:            tc.networkRange,
					PrivateSubnetMaskBits:   tc.privateSubnetMaskBits,
					PublicSubnetMaskBits:    tc.publicSubnetMaskBits,
				}

				newResource, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			err = newResource.EnsureCreated(context.Background(), tc.object)

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
		})
	}
}

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
}
