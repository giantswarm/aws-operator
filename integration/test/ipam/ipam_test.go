// +build k8srequired

package ipam

import (
	"context"
	"testing"
)

func Test_IPAM(t *testing.T) {
	err := i.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
