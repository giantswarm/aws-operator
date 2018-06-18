// +build k8srequired

package masternode

import (
	"context"
	"testing"
)

func Test_Master_Node(t *testing.T) {
	err := m.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
