package ipam

import (
	"context"
	"net"
	"reflect"

	"github.com/giantswarm/microerror"
)

type TestPersister struct {
	subnet net.IPNet
}

func NewTestPersister(subnet net.IPNet) *TestPersister {
	p := &TestPersister{
		subnet: subnet,
	}

	return p
}

func (p *TestPersister) Persist(ctx context.Context, subnet net.IPNet, namespace string, name string) error {
	if !reflect.DeepEqual(subnet, p.subnet) {
		return microerror.Mask(invalidConfigError)
	}

	return nil
}
