// +build k8srequired

package scaling

import (
	"context"
	"testing"
)

func Test_Scaling_Workers(t *testing.T) {
	err := s.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
