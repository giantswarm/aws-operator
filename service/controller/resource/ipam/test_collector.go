package ipam

import (
	"context"
	"net"
)

type TestCollector struct {
	subnets []net.IPNet
}

func NewTestCollector(subnets []net.IPNet) *TestCollector {
	c := &TestCollector{
		subnets: subnets,
	}

	return c
}

func (c *TestCollector) Collect(ctx context.Context, networkRange net.IPNet) ([]net.IPNet, error) {
	return c.subnets, nil
}
