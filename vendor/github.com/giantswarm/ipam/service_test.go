package ipam

import (
	"bytes"
	"net"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/memory"
)

// TestNew tests the New function.
func TestNew(t *testing.T) {
	testLogger := microloggertest.New()
	testStorage, err := memory.New(memory.DefaultConfig())
	if err != nil {
		t.Fatalf("error creating new storage: %v", err)
	}

	_, testNetwork, _ := net.ParseCIDR("10.4.0.0/16")

	tests := []struct {
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		// Test that the default config, with a network set,
		// returns a new IPAM service.
		{
			config: func() Config {
				c := DefaultConfig()

				c.Logger = testLogger
				c.Storage = testStorage
				c.Network = testNetwork

				return c
			},
		},

		// Test that a config with a nil logger returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()

				c.Logger = nil
				c.Storage = testStorage
				c.Network = testNetwork

				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with a nil storage returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()

				c.Logger = testLogger
				c.Storage = nil
				c.Network = testNetwork

				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with an empty network returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()

				c.Logger = testLogger
				c.Storage = testStorage
				c.Network = nil

				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},
	}

	for index, test := range tests {
		service, err := New(test.config())

		if err == nil && test.expectedErrorHandler != nil {
			t.Fatalf("%v: expected error not returned", index)
		}
		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned: %v", index, err)
			} else {
				if !test.expectedErrorHandler(err) {
					t.Fatalf("%v: incorrect error returned: %v", index, err)
				}
			}
		} else {
			if service == nil {
				t.Fatalf("%v: service is nil", index)
			}
		}
	}
}

// TestNewSubnetAndDeleteSubnet tests that NewSubnet and DeleteSubnet methods work together correctly.
func TestNewSubnetAndDeleteSubnet(t *testing.T) {
	type step struct {
		// add is true if we create a subnet, false if we delete one.
		add bool
		// mask is the mask to use if creating a new subnet.
		mask int
		// subnetToDelete is the subnet to delete, if we are deleting one.
		subnetToDelete string
		// expectedSubnet is the subnet we expect, if we are creating one.
		expectedSubnet string
		// expectedErrorHandler is the error handler we expect in case of error,
		// if an error should be returned.
		expectedErrorHandler func(error) bool
	}

	tests := []struct {
		network string
		steps   []step
	}{
		// Test that adding a single subnet returns the correct subnet.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
			},
		},

		// Test that adding three subnets returns the correct subnets.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.1.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.2.0/24",
				},
			},
		},

		// Test that adding two subnets with different mask sizes,
		// returns the correct subnets.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           25,
					expectedSubnet: "10.4.0.0/25",
				},
				{
					add:            true,
					mask:           26,
					expectedSubnet: "10.4.0.128/26",
				},
			},
		},

		// Test adding a network that is too large returns an error.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:                  true,
					mask:                 15,
					expectedErrorHandler: IsMaskTooBig,
				},
			},
		},

		// Test that adding a subnet, deleting it, and adding another subnet,
		// works correctly.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
				{
					add:            false,
					subnetToDelete: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
			},
		},

		// Test that adding two subnets, deleting the first one, then
		// adding a third larger subnet, and then a fourth of the original size,
		// works correctly.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.1.0/24",
				},
				{
					add:            false,
					subnetToDelete: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           23,
					expectedSubnet: "10.4.2.0/23",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
			},
		},
	}

	for index, test := range tests {
		// Parse network.
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
		}

		// Create a new IPAM service.
		logger := microloggertest.New()
		storage, err := memory.New(memory.DefaultConfig())
		if err != nil {
			t.Fatalf("%v: error creating new storage: %v", index, err)
		}

		config := DefaultConfig()
		config.Logger = logger
		config.Storage = storage
		config.Network = network

		service, err := New(config)
		if err != nil {
			t.Fatalf("%v: error returned creating ipam service: %v", index, err)
		}

		for _, step := range test.steps {
			if step.add {
				mask := net.CIDRMask(step.mask, 32)

				returnedSubnet, err := service.NewSubnet(mask)
				if err != nil {
					if step.expectedErrorHandler != nil {
						if !step.expectedErrorHandler(err) {
							t.Fatalf("%v: incorrect error returned creating new subnet: %v", index, err)
						}
					} else {
						t.Fatalf("%v: unexpected error returned creating new subnet: %v", index, err)
					}
				} else {
					_, expectedSubnet, err := net.ParseCIDR(step.expectedSubnet)
					if err != nil {
						t.Fatalf("%v: error returned parsing expected subnet: %v", index, err)
					}

					if !returnedSubnet.IP.Equal(expectedSubnet.IP) || !bytes.Equal(returnedSubnet.Mask, expectedSubnet.Mask) {
						t.Fatalf(
							"%v: returned subnet did not match expected.\nexpected: %v\nreturned: %v\n",
							index,
							*expectedSubnet,
							returnedSubnet,
						)
					}
				}
			} else {
				_, subnetToDelete, err := net.ParseCIDR(step.subnetToDelete)
				if err != nil {
					t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
				}

				if err := service.DeleteSubnet(*subnetToDelete); err != nil {
					if !step.expectedErrorHandler(err) {
						t.Fatalf("%v: unexpected error returned creating new subnet: %v", index, err)
					}
				}
			}
		}
	}
}

// TestNewWithAllocatedNetworks tests that the NewSubnet method works correctly
// with the allocated networks configuration.
func TestNewWithAllocatedNetworks(t *testing.T) {
	tests := []struct {
		network          string
		allocatedSubnets []string
		mask             int

		expectedErrorHandler func(error) bool
		expectedSubnet       string
	}{
		// Test that a nil allocated subnets slice does not affect subnet handout.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: nil,
			mask:             24,

			expectedSubnet: "10.0.0.0/24",
		},

		// Test that an empty allocated subnets slice does not affect subnet handout.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: []string{},
			mask:             24,

			expectedSubnet: "10.0.0.0/24",
		},

		// Test that an allocated subnet, with a mask greater than the network mask,
		// blocks the start of the range, and that the next subnet starts after it.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: []string{"10.0.0.0/17"},
			mask:             24,

			expectedSubnet: "10.0.128.0/24",
		},

		// Test that two allocated subnets, with masks greater than the network mask,
		// block the start of the range, and the next subnet starts after them.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: []string{"10.0.0.0/24", "10.0.1.0/24"},
			mask:             24,

			expectedSubnet: "10.0.2.0/24",
		},

		// Test that an allocated subnet, after the start of the range,
		// does not block the start of the range.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: []string{"10.0.1.0/24"},
			mask:             24,

			expectedSubnet: "10.0.0.0/24",
		},

		// Test that an error is returned when creating the IPAM service if
		// the allocated subnet is too big.
		{
			network:          "10.0.0.0/24",
			allocatedSubnets: []string{"10.0.0.0/23"},

			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that an error is returned when creating the IPAM service
		// if any of the allocated subnets are too large.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: []string{"10.0.0.0/24", "10.0.0.0/8"},

			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that an error is returned when creating the IPAM service
		// if an allocated subnet does not belong to the network.
		{
			network:          "10.0.0.0/16",
			allocatedSubnets: []string{"192.168.0.0/16"},

			expectedErrorHandler: IsInvalidConfig,
		},
	}

	for index, test := range tests {
		logger := microloggertest.New()
		storage, err := memory.New(memory.DefaultConfig())
		if err != nil {
			t.Fatalf("%v: error creating new storage: %v", index, err)
		}

		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
		}

		var allocatedSubnets []net.IPNet
		if test.allocatedSubnets == nil {
			allocatedSubnets = nil
		} else {
			for _, allocatedSubnet := range test.allocatedSubnets {
				_, subnet, err := net.ParseCIDR(allocatedSubnet)
				if err != nil {
					t.Fatalf("%v: error returned parsing subnet cidr: %v", index, err)
				}
				allocatedSubnets = append(allocatedSubnets, *subnet)
			}
		}

		config := DefaultConfig()
		config.Logger = logger
		config.Storage = storage

		config.AllocatedSubnets = allocatedSubnets
		config.Network = network

		service, err := New(config)
		if err != nil && test.expectedErrorHandler == nil {
			t.Fatalf("%v: unexpected error returned creating ipam service: %v\n", index, err)
		}
		if err != nil && !test.expectedErrorHandler(err) {
			t.Fatalf("%v: incorrect error returned creating ipam service: %v\n", index, err)
		}
		if err == nil && test.expectedErrorHandler != nil {
			t.Fatalf("%v: expected error not returned creating ipam service\n", index)
		}

		if test.expectedSubnet != "" {
			_, expectedSubnet, err := net.ParseCIDR(test.expectedSubnet)
			if err != nil {
				t.Fatalf("%v: error returned parsing expected subnet: %v", index, err)
			}

			returnedSubnet, err := service.NewSubnet(net.CIDRMask(test.mask, 32))
			if err != nil {
				t.Fatalf("%v: error returned creating new subnet: %v\n", index, err)
			}

			if !returnedSubnet.IP.Equal(expectedSubnet.IP) || !bytes.Equal(returnedSubnet.Mask, expectedSubnet.Mask) {
				t.Fatalf(
					"%v: returned subnet did not match expected.\nexpected: %v\nreturned: %v\n",
					index,
					*expectedSubnet,
					returnedSubnet,
				)
			}
		}
	}
}
