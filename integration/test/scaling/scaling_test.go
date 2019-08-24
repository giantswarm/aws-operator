package scaling

import (
	"context"
	"testing"
)

func Test_Scaling_Workers(t *testing.T) {
	err := scalingTest.Test(context.Background())
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
