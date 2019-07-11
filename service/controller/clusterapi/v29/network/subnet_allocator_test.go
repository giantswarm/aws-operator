package network

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
)

var errArtificial = errors.New("artificial error")

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
}

func Test_SubnetAllocator(t *testing.T) {
	testCases := []struct {
		name           string
		callbacks      Callbacks
		networkRange   net.IPNet
		subnetSize     net.IPMask
		expectedSubnet net.IPNet
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: allocate first subnet",
			callbacks: Callbacks{
				Collect: func(_ context.Context) ([]net.IPNet, error) { return []net.IPNet{}, nil },
				Persist: func(_ context.Context, _ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: mustParseCIDR("10.100.0.0/24"),
			errorMatcher:   nil,
		},
		{
			name: "case 1: allocate fourth subnet",
			callbacks: Callbacks{
				Collect: func(_ context.Context) ([]net.IPNet, error) {
					reservedNetworks := []net.IPNet{
						mustParseCIDR("10.100.0.0/24"),
						mustParseCIDR("10.100.1.0/24"),
						mustParseCIDR("10.100.3.0/24"),
					}
					return reservedNetworks, nil
				},
				Persist: func(_ context.Context, _ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: mustParseCIDR("10.100.2.0/24"),
			errorMatcher:   nil,
		},
		{
			name: "case 2: handle error from getting reserved networks",
			callbacks: Callbacks{
				Collect: func(_ context.Context) ([]net.IPNet, error) { return nil, errArtificial },
				Persist: func(_ context.Context, _ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: net.IPNet{},
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == errArtificial },
		},
		{
			name: "case 3: handle error from persisting allocated network",
			callbacks: Callbacks{
				Collect: func(_ context.Context) ([]net.IPNet, error) { return []net.IPNet{}, nil },
				Persist: func(_ context.Context, _ net.IPNet) error { return errArtificial },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: net.IPNet{},
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == errArtificial },
		},
	}

	svc, err := NewSubnetAllocator(SubnetAllocatorConfig{Logger: microloggertest.New()})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			net, err := svc.Allocate(context.Background(), tc.networkRange, tc.subnetSize, tc.callbacks)

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
