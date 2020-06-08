package tccpf

import (
	"errors"
	"testing"

	"github.com/giantswarm/microerror"
)

func Test_IsNoUpdate(t *testing.T) {
	testCases := []struct {
		name  string
		err   string
		match bool
	}{
		{
			name:  "case 0",
			err:   "An error occurred (ValidationError) when calling the UpdateStack operation: No updates are to be performed.",
			match: true,
		},
		{
			name:  "case 1",
			err:   "ValidationError: No updates are to be performed.",
			match: true,
		},
		{
			name:  "case 2",
			err:   "ValidationError: no update to be performed",
			match: true,
		},
		{
			name:  "case 3",
			err:   "not found error",
			match: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNoUpdate(microerror.Mask(errors.New(tc.err)))

			if result != tc.match {
				t.Fatalf("expected %t, got %t", tc.match, result)
			}
		})
	}
}
