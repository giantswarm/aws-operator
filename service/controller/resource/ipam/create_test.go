package ipam

import (
	"context"
	"net"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/internal/locker"
	"github.com/giantswarm/aws-operator/service/internal/unittest"
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
		// If the network pool CIDR is given the test simulates the injection of
		// a user configured custom network range.
		networkPool net.IPNet
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
			networkPool:             net.IPNet{},
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
			networkPool:             net.IPNet{},
		},
		{
			name: "case 2 allocate first subnet via network pool",

			checker:   NewTestChecker(true),
			collector: NewTestCollector([]net.IPNet{}),
			persister: NewTestPersister(mustParseCIDR("10.100.0.0/24")),

			allocatedSubnetMaskBits: 24,
			networkRange:            mustParseCIDR("127.0.0.1/8"), // dummy we ignore in the test
			privateSubnetMaskBits:   25,
			publicSubnetMaskBits:    25,
			networkPool:             mustParseCIDR("10.100.0.0/16"),
		},
		{
			name: "case 3 allocate fourth subnet via network pool",

			checker: NewTestChecker(true),
			collector: NewTestCollector([]net.IPNet{
				mustParseCIDR("10.100.0.0/24"),
				mustParseCIDR("10.100.1.0/24"),
				mustParseCIDR("10.100.3.0/24"),
			}),
			persister: NewTestPersister(mustParseCIDR("10.100.2.0/24")),

			allocatedSubnetMaskBits: 24,
			networkRange:            mustParseCIDR("127.0.0.1/8"), // dummy we ignore in the test
			privateSubnetMaskBits:   25,
			publicSubnetMaskBits:    25,
			networkPool:             mustParseCIDR("10.100.0.0/16"),
		},
		{
			name: "case 4 allocate with multiple subnet sizes",

			checker: NewTestChecker(true),
			collector: NewTestCollector([]net.IPNet{
				mustParseCIDR("10.163.0.0/19"),
				mustParseCIDR("10.163.30.0/23"),
				mustParseCIDR("10.163.30.128/24"),
				mustParseCIDR("10.163.31.0/24"),
				mustParseCIDR("10.163.32.0/21"),
				mustParseCIDR("10.163.32.0/24"),
				mustParseCIDR("10.163.33.0/24"),
				mustParseCIDR("10.163.34.0/24"),
				mustParseCIDR("10.163.35.0/24"),
				mustParseCIDR("10.163.36.0/24"),
				mustParseCIDR("10.163.37.0/24"),
				mustParseCIDR("10.163.40.0/24"),
			}),
			persister: NewTestPersister(mustParseCIDR("10.163.42.0/23")),

			allocatedSubnetMaskBits: 23,
			networkRange:            mustParseCIDR("10.161.0.0/16"), // dummy we ignore in the test
			privateSubnetMaskBits:   24,
			publicSubnetMaskBits:    24,
			networkPool:             mustParseCIDR("10.163.0.0/16"),
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContextControlPlane()
			k := unittest.FakeK8sClient()

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

			var cr infrastructurev1alpha3.AWSCluster
			{
				cr = unittest.DefaultCluster()

				if !netIPEmpty(tc.networkPool) {
					cr.Spec.Provider.Nodes.NetworkPool = cr.GetName()
				}

				err = k.CtrlClient().Create(ctx, &cr)
				if err != nil {
					t.Fatal(err)
				}
			}

			if !netIPEmpty(tc.networkPool) {
				cr := unittest.DefaultNetworkPool(tc.networkPool.String())

				err = k.CtrlClient().Create(ctx, &cr)
				if err != nil {
					t.Fatal(err)
				}
			}

			var newResource *Resource
			{
				c := Config{
					Checker:   tc.checker,
					Collector: tc.collector,
					K8sClient: k,
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

			err = newResource.EnsureCreated(context.Background(), &cr)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func netIPEmpty(netip net.IPNet) bool {
	return netip.String() == "<nil>"
}

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
}
