package key

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSanitizeCFResourceName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "case 0: simple string with dash",
			input:    "abc-123",
			expected: "Abc123",
		},
		{
			name:     "case 1: unicode story",
			input:    "Dear god why? щ（ﾟДﾟщ）",
			expected: "DearGodWhy",
		},
		{
			name:     "case 2: AWS AZ",
			input:    "foo-bar-eu-central-1b",
			expected: "FooBarEuCentral1b",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := SanitizeCFResourceName(tc.input)

			if output != tc.expected {
				t.Fatalf("\n\n%s\n", cmp.Diff(output, tc.expected))
			}
		})
	}
}
