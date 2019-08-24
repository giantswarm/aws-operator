package ipam

import (
	"context"
	"testing"
)

func Test_IPAM(t *testing.T) {
	err := ipamTest.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
