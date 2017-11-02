package versionbundle

import (
	"testing"
)

func Test_Changelog_Validate(t *testing.T) {
	testCases := []struct {
		Changelog    Changelog
		ErrorMatcher func(err error) bool
	}{
		// Test 0 ensures an empty changelog is not valid.
		{
			Changelog:    Changelog{},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 1 is the same as 0 but with initialized properties.
		{
			Changelog: Changelog{
				Component:   "",
				Description: "",
				Kind:        "",
			},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 2 ensures a changelog missing description and kind throws an error.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "",
				Kind:        "",
			},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 2 ensures a changelog missing kind throws an error.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "",
			},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 3 ensures a changelog missing a valid kind throws an error.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "foo",
			},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 4 ensures a changelog missing a description throws an error.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "",
				Kind:        "added",
			},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 5 ensures a changelog missing a component throws an error.
		{
			Changelog: Changelog{
				Component:   "",
				Description: "description",
				Kind:        "added",
			},
			ErrorMatcher: IsInvalidChangelog,
		},

		// Test 6 ensures a changelog having a valid kind does not throw an error.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "added",
			},
			ErrorMatcher: nil,
		},

		// Test 7 is the same as 6 but with a different kind.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "changed",
			},
			ErrorMatcher: nil,
		},

		// Test 8 is the same as 6 but with a different kind.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "deprecated",
			},
			ErrorMatcher: nil,
		},

		// Test 9 is the same as 6 but with a different kind.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "fixed",
			},
			ErrorMatcher: nil,
		},

		// Test 10 is the same as 6 but with a different kind.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "removed",
			},
			ErrorMatcher: nil,
		},

		// Test 11 is the same as 6 but with a different kind.
		{
			Changelog: Changelog{
				Component:   "kubernetes",
				Description: "description",
				Kind:        "security",
			},
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		err := tc.Changelog.Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}
