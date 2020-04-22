package drainerfinalizer

import (
	"errors"
	"testing"

	"github.com/giantswarm/microerror"
)

func Test_IsNoActiveLifeCycleAction(t *testing.T) {
	testCases := []struct {
		name  string
		err   string
		match bool
	}{
		{
			name:  "case 0",
			err:   "ValidationError: No active Lifecycle Action found with instance ID i-08406e13ee788fc10",
			match: true,
		},
		{
			name:  "case 1",
			err:   "ValidationError: no active lifecycle action found with instance id i-08406e13ee788fc10",
			match: true,
		},
		{
			name:  "case 2",
			err:   "not found error",
			match: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNoActiveLifeCycleAction(microerror.Mask(errors.New(tc.err)))

			if result != tc.match {
				t.Fatalf("expected %t, got %t", tc.match, result)
			}
		})
	}
}
