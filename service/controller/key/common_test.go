package key

import (
	"strconv"
	"testing"

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
		releaseVersion string
		expected       bool
	}{
		{
			name:           "empty version",
			flatcarVersion: "",
			releaseVersion: "3510.2.6",
			expected:       false,
		},
		{
			name:           "same version",
			flatcarVersion: "3510.2.6",
			releaseVersion: "3510.2.6",
			expected:       false,
		},
		{
			name:           "lower version",
			flatcarVersion: "3488.0.0",
			releaseVersion: "3510.2.6",
			expected:       false,
		},
		{
			name:           "higher version",
			flatcarVersion: "3690.0.0",
			releaseVersion: "3510.2.6",
			expected:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsFlatcarVersionNewer(tc.releaseVersion, tc.flatcarVersion)
			if result != tc.expected {
				t.Errorf("expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
