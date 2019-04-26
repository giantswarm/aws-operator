package ipam

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

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

func Test_NetworkAllocator(t *testing.T) {
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
				GetReservedNetworks:     func() ([]net.IPNet, error) { return []net.IPNet{}, nil },
				PersistAllocatedNetwork: func(_ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: mustParseCIDR("10.100.0.0/24"),
			errorMatcher:   nil,
		},
		{
			name: "case 1: allocate fourth subnet",
			callbacks: AllocationCallbacks{
				GetReservedNetworks: func() ([]net.IPNet, error) {
					reservedNetworks := []net.IPNet{
						mustParseCIDR("10.100.0.0/24"),
						mustParseCIDR("10.100.1.0/24"),
						mustParseCIDR("10.100.3.0/24"),
					}
					return reservedNetworks, nil
				},
				PersistAllocatedNetwork: func(_ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: mustParseCIDR("10.100.2.0/24"),
			errorMatcher:   nil,
		},
		{
			name: "case 2: handle error from getting reserved networks",
			callbacks: AllocationCallbacks{
				GetReservedNetworks:     func() ([]net.IPNet, error) { return nil, errArtificial },
				PersistAllocatedNetwork: func(_ net.IPNet) error { return nil },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: net.IPNet{},
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == errArtificial },
		},
		{
			name: "case 3: handle error from persisting allocated network",
			callbacks: AllocationCallbacks{
				GetReservedNetworks:     func() ([]net.IPNet, error) { return []net.IPNet{}, nil },
				PersistAllocatedNetwork: func(_ net.IPNet) error { return errArtificial },
			},
			networkRange:   mustParseCIDR("10.100.0.0/16"),
			subnetSize:     net.CIDRMask(24, 32),
			expectedSubnet: net.IPNet{},
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == errArtificial },
		},
	}

	svc, err := New(Config{Logger: microloggertest.New()})
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

func Test_NetworkAllocator_Locking(t *testing.T) {
	svc, err := New(Config{Logger: microloggertest.New()})
	if err != nil {
		t.Fatal(err)
	}

	/*
		This is how this test is expected to execute:
		  * Create a channel that is used to signal from thread #1 when it is
		    inside a lock in Allocate().
		  * In thread #2 wait for that signal before calling Allocate().
		  * When thread #1 has sent signal, it sleeps for a little while so that it
		    guarantees that thread #2 is waiting for mutex.
		  * Each thread then performs subnet allocation and verifies that allocated
		    subnet matches the expectation.
	*/

	fullRange := mustParseCIDR("10.100.0.0/16")
	netSize := net.CIDRMask(24, 32)

	var reservedNetworks *[]net.IPNet
	{
		// Take a pointer to slice so that behaviour is correct between
		// goroutines during re-allocations in slice.
		slice := make([]net.IPNet, 0)
		reservedNetworks = &slice
	}

	// signal is the channel that is used from thread #2 to signal that thread
	// #1 can call Allocate().
	signal := make(chan struct{})

	// wg is a WaitGroup for this test to wait until both threads have
	// executed.
	wg := &sync.WaitGroup{}

	// Thread #1
	wg.Add(1)
	go func() {
		defer wg.Done()

		callbacks := AllocationCallbacks{
			GetReservedNetworks: func() ([]net.IPNet, error) {
				// Allow second thread to call AllocateNetwork.
				signal <- struct{}{}

				// Add a bit of delay to let it catch up.
				time.Sleep(100 * time.Millisecond)

				return *reservedNetworks, nil
			},
			PersistAllocatedNetwork: func(n net.IPNet) error {
				*reservedNetworks = append(*reservedNetworks, n)
				return nil
			},
		}

		numReservedNetworks := len(*reservedNetworks)
		numExpectedReservedNetworks := 0
		if numReservedNetworks != numExpectedReservedNetworks {
			t.Errorf("expected len(reservedNetworks) == %d, got %d", numExpectedReservedNetworks, numReservedNetworks)
		}
		_, err := svc.Allocate(context.Background(), fullRange, netSize, callbacks)
		if err != nil {
			t.Error(err)
		}

		numReservedNetworks = len(*reservedNetworks)
		numExpectedReservedNetworks = 1
		if numReservedNetworks != numExpectedReservedNetworks {
			t.Errorf("expected len(reservedNetworks) == %d, got %d", numExpectedReservedNetworks, numReservedNetworks)
		}

		expectedNetwork := mustParseCIDR("10.100.0.0/24")
		gotNetwork := (*reservedNetworks)[0]
		if !reflect.DeepEqual(gotNetwork, expectedNetwork) {
			t.Errorf("expected subnet %q to be allocated, got %q", expectedNetwork.String(), gotNetwork.String())
		}
	}()

	// Thread #2
	wg.Add(1)
	go func() {
		defer wg.Done()

		callbacks := AllocationCallbacks{
			GetReservedNetworks: func() ([]net.IPNet, error) {
				fmt.Printf("reservedNetworks: %#v\n", *reservedNetworks)
				return *reservedNetworks, nil
			},
			PersistAllocatedNetwork: func(n net.IPNet) error {
				*reservedNetworks = append(*reservedNetworks, n)
				return nil
			},
		}

		// Wait for signal from first thread before getting into allocation.
		<-signal

		_, err := svc.Allocate(context.Background(), fullRange, netSize, callbacks)
		if err != nil {
			t.Error(err)
		}

		numReservedNetworks := len(*reservedNetworks)
		numExpectedReservedNetworks := 2
		if numReservedNetworks != numExpectedReservedNetworks {
			t.Errorf("expected len(reservedNetworks) == %d, got %d", numExpectedReservedNetworks, numReservedNetworks)
		}

		expectedNetwork := mustParseCIDR("10.100.1.0/24")
		gotNetwork := (*reservedNetworks)[1]
		if !reflect.DeepEqual(gotNetwork, expectedNetwork) {
			t.Errorf("expected subnet %q to be allocated, got %q", expectedNetwork.String(), gotNetwork.String())
		}
	}()

	wg.Wait()
}
