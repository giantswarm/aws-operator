package key

import (
	"strconv"
	"testing"

	"github.com/blang/semver"
	"github.com/google/go-cmp/cmp"
)

func TestSanitizeCFResourceName(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "case 0: simple string with dash",
			input: []string{
				"abc-123",
			},
			expected: "Abc123",
		},
		{
			name: "case 1: unicode story",
			input: []string{
				"Dear god why? щ（ﾟДﾟщ）",
			},
			expected: "DearGodWhy",
		},
		{
			name: "case 2: AWS AZ",
			input: []string{
				"foo-bar-eu-central-1b",
			},
			expected: "FooBarEuCentral1b",
		},
		{
			name: "case 3: multiple inputs",
			input: []string{
				"foo-bar-eu-central-1b",
				"abc-123",
			},
			expected: "FooBarEuCentral1bAbc123",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := SanitizeCFResourceName(tc.input...)

			if output != tc.expected {
				t.Fatalf("\n\n%s\n", cmp.Diff(output, tc.expected))
			}
		})
	}
}

func TestIsMinimumFlatcarVersion(t *testing.T) {
	testCases := []struct {
		name           string
		flatcarVersion string
		expected       bool
	}{
		{
			name:           "empty version",
			flatcarVersion: "",
			expected:       false,
		},
		{
			name:           "wrong version",
			flatcarVersion: "33",
			expected:       false,
		},
		{
			name:           "minimum version",
			flatcarVersion: "3689.0.0",
			expected:       true,
		},
		{
			name:           "lower version",
			flatcarVersion: "3688.0.0",
			expected:       false,
		},
		{
			name:           "higher version",
			flatcarVersion: "3690.0.0",
			expected:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version, _ := semver.New(tc.flatcarVersion)
			result := IsMinimumFlatcarVersion(version)
			if result != tc.expected {
				t.Errorf("expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
