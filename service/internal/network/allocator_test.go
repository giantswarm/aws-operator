package network

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/internal/locker"
)

var errArtificial = errors.New("artificial error")

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
}

func Test_Allocator(t *testing.T) {
	testCases := []struct {
		name           string
		callbacks      AllocationCallbacks
		networkRange   net.IPNet
		subnetSize     net.IPMask
		expectedSubnet net.IPNet
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: allocate first subnet",
			callbacks: AllocationCallbacks{
				GetReservedNetworks:     func(_ context.Context) ([]net.IPNet, error) { return []net.IPNet{}, nil },
				PersistAllocatedNetwork: func(_ context.Context, _ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: mustParseCIDR("10.100.0.0/24"),
			errorMatcher:   nil,
		},
		{
			name: "case 1: allocate fourth subnet",
			callbacks: AllocationCallbacks{
				GetReservedNetworks: func(_ context.Context) ([]net.IPNet, error) {
					reservedNetworks := []net.IPNet{
						mustParseCIDR("10.100.0.0/24"),
						mustParseCIDR("10.100.1.0/24"),
						mustParseCIDR("10.100.3.0/24"),
					}
					return reservedNetworks, nil
				},
				PersistAllocatedNetwork: func(_ context.Context, _ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: mustParseCIDR("10.100.2.0/24"),
			errorMatcher:   nil,
		},
		{
			name: "case 2: handle error from getting reserved networks",
			callbacks: AllocationCallbacks{
				GetReservedNetworks:     func(_ context.Context) ([]net.IPNet, error) { return nil, errArtificial },
				PersistAllocatedNetwork: func(_ context.Context, _ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: net.IPNet{},
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == errArtificial },
		},
		{
			name: "case 3: handle error from persisting allocated network",
			callbacks: AllocationCallbacks{
				GetReservedNetworks:     func(_ context.Context) ([]net.IPNet, error) { return []net.IPNet{}, nil },
				PersistAllocatedNetwork: func(_ context.Context, _ net.IPNet) error { return errArtificial },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: net.IPNet{},
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == errArtificial },
		},
	}

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

	var a Allocator
	{
		c := Config{
			Locker: mutexLocker,
			Logger: microloggertest.New(),
		}

		a, err = New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			net, err := a.Allocate(context.Background(), tc.networkRange, tc.subnetSize, tc.callbacks)

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

			if !reflect.DeepEqual(net, tc.expectedSubnet) {
				t.Fatalf("Allocated subnet == %q, want %q", net.String(), tc.expectedSubnet.String())
			}
		})
	}
}
