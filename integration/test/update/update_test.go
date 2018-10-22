// +build k8srequired

package update

import (
	"context"
	"testing"
)

func Test_Update(t *testing.T) {
	err := updateTest.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
