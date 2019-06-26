// +build k8srequired

package loadtest

import (
	"context"
	"testing"
)

func Test_Load_Test(t *testing.T) {
	err := loadTestTest.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
