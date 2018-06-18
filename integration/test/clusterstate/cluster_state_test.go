// +build k8srequired

package clusterstate

import (
	"context"
	"testing"
)

func Test_Master_Node(t *testing.T) {
	err := cs.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
