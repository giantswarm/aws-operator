package ipam

import (
	"context"
	"net"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/internal/locker"
)

func Test_SubnetAllocator(t *testing.T) {
	testCases := []struct {
		name string

		checker   Checker
		collector Collector
		persister Persister

		allocatedSubnetMaskBits int
		networkRange            net.IPNet
		privateSubnetMaskBits   int
		publicSubnetMaskBits    int
	}{
		{
			name: "case 0 allocate first subnet",

			checker:   NewTestChecker(true),
			collector: NewTestCollector([]net.IPNet{}),
			persister: NewTestPersister(mustParseCIDR("10.100.0.0/24")),

			allocatedSubnetMaskBits: 24,
			networkRange:            mustParseCIDR("10.100.0.0/16"),
			privateSubnetMaskBits:   25,
			publicSubnetMaskBits:    25,
		},
		{
			name: "case 1 allocate fourth subnet",

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

			var mutexLocker locker.Interface
			{
				c := locker.MutexLockerConfig{
					Logger: microloggertest.New(),
				}

				mutexLocker, err = locker.NewMutexLocker(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var newResource *Resource
			{
				c := Config{
					Checker:   tc.checker,
					Collector: tc.collector,
					Locker:    mutexLocker,
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

			err = newResource.EnsureCreated(context.Background(), &infrastructurev1alpha2.AWSCluster{})
			if err != nil {
				t.Fatal(err)
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
