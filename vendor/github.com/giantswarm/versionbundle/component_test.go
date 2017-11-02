package versionbundle

import (
	"testing"
)

func Test_Component_Validate(t *testing.T) {
	testCases := []struct {
		Component    Component
		ErrorMatcher func(err error) bool
	}{
		// Test 0 ensures an empty component is not valid.
		{
			Component:    Component{},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 1 is the same as 0 but with initialized properties.
		{
			Component: Component{
				Name:    "",
				Version: "",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 2 ensures a component missing a version is not valid.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 3 ensures a non-semver version is invalid.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "foo",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 4 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 5 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 6 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.x",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 7 is the same as 11 but with a different version.
		//
		// NOTE using the wildcard for minor does not allow a wildcard for patch.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.x.2",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 8 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.x.x.x",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 9 ensures using the wildcard for the major version is not valid.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "x.x.x",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 10 is the same as 9 but with different input.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "x.3.9",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 11 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.3.9.4",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 12 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.3.foo",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 13 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.bar.3",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 14 is the same as 11 but with a different version.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "kubernetes.14.3",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 15 is the same as 11 but with a different version.
		//
		// NOTE negative version numbers are not allowed.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.-3.7",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 16 is the same as 13 but with a different version.
		//
		// NOTE negative version numbers are not allowed.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "-1.3.7",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 17 is the same as 13 but with a different version.
		//
		// NOTE negative version numbers are not allowed.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.3.-7",
			},
			ErrorMatcher: IsInvalidComponent,
		},

		// Test 18 ensures a valid component does not throw an error.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "1.3.9",
			},
			ErrorMatcher: nil,
		},

		// Test 19 ensures a valid component does not throw an error.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "0.0.0",
			},
			ErrorMatcher: nil,
		},

		// Test 20 ensures a valid component does not throw an error.
		{
			Component: Component{
				Name:    "kubernetes",
				Version: "11.3.785",
			},
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		err := tc.Component.Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}
