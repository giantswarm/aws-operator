package cloudconfig

import (
	"testing"
)

func Test_Template(t *testing.T) {
	testCases := []struct {
		Version      version
		ErrorMatcher func(err error) bool
	}{
		{
			Version:      V_0_1_0,
			ErrorMatcher: nil,
		},
		{
			Version:      version("foo"),
			ErrorMatcher: IsNotFound,
		},
	}

	for i, tc := range testCases {
		template, err := NewTemplate(tc.Version)
		if err != nil {
			if tc.ErrorMatcher != nil && !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i+1, true, false)
			}
		} else {
			if template.Master == "" {
				t.Fatalf("test %d expected %#v got %#v", i+1, "valid master template", "empty string")
			}

			if template.Worker == "" {
				t.Fatalf("test %d expected %#v got %#v", i+1, "valid worker template", "empty string")
			}
		}
	}
}
