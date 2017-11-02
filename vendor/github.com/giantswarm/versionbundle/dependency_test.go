package versionbundle

import (
	"testing"
)

func Test_Dependency_Matches(t *testing.T) {
	testCases := []struct {
		Dependency     Dependency
		Component      Component
		ExpectedResult bool
	}{
		// Test 0 ensures when using operator '==' the same version of dependency
		// and component matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 1 ensures when using operator '>=' the same version of dependency
		// and component matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 2 ensures when using operator '<=' the same version of dependency
		// and component matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "<= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 3 ensures when using operator '<' the same version of dependency
		// and component does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 4 ensures when using operator '>' the same version of dependency
		// and component does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 5 ensures when using operator '!=' the same version of dependency
		// and component does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 6 ensures when using operator '==' a higher component patch version
		// does not match with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: false,
		},

		// Test 7 ensures when using operator '>=' a higher component patch version
		// macthes with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: true,
		},

		// Test 8 ensures when using operator '>' a higher component patch version
		// matches with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: true,
		},

		// Test 9 ensures when using operator '<' a higher component patch version
		// does not match with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: false,
		},

		// Test 10 ensures when using operator '!=' a higher component patch version
		// matchtes with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: true,
		},

		// Test 11 ensures when using operator '==' a lower component patch version
		// does not match with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: false,
		},

		// Test 12 ensures when using operator '>=' a lower component patch version
		// does not match with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: false,
		},

		// Test 13 ensures when using operator '>' a lower component patch version
		// does not match with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: false,
		},

		// Test 14 ensures when using operator '<' a lower component patch version
		// matches with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: true,
		},

		// Test 15 ensures when using operator '!=' a lower component patch version
		// matches with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.7.1",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: true,
		},

		// Test 16 ensures when using operator '==' and a wildcard for the
		// dependency patch version the same version of dependency and component
		// matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 17 ensures when using operator '>=' the same version of dependency
		// and component matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 18 ensures when using operator '<=' and a wildcard for the
		// dependency patch version the same version of dependency and component
		// matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "<= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 19 ensures when using operator '<' and a wildcard for the dependency
		// patch version the same version of dependency and component does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 20 ensures when using operator '>' and a wildcard for the dependency
		// patch version the same version of dependency and component does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 21 ensures when using operator '!=' and a wildcard for the dependency
		// patch version the same version of dependency and component does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 22 ensures when using operator '>=' and a wildcard for the
		// dependency patch version a higher component patch version does match with
		// the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: true,
		},

		// Test 23 ensures when using operator '>' and a wildcard for the dependency
		// patch version a higher component patch version does not match with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: false,
		},

		// Test 24 ensures when using operator '<' and a wildcard for the dependency
		// patch version a higher component patch version does not match with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: false,
		},

		// Test 25 ensures when using operator '!=' and a wildcard for the
		// dependency patch version a higher component patch version does not match
		// with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.2",
			},
			ExpectedResult: false,
		},

		// Test 26 ensures when using operator '==' and a wildcard for the
		// dependency patch version a lower component patch version matchtes with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: true,
		},

		// Test 27 ensures when using operator '>=' and a wildcard for the
		// dependency patch version a lower component patch version matchtes with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: true,
		},

		// Test 28 ensures when using operator '>' and a wildcard for the dependency
		// patch version a lower component patch version does not match with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: false,
		},

		// Test 29 ensures when using operator '<' and a wildcard for the dependency
		// patch version a lower component patch version does not match with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: false,
		},

		// Test 30 ensures when using operator '!=' and a wildcard for the
		// dependency patch version a lower component patch version does not match
		// with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.7.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.0",
			},
			ExpectedResult: false,
		},

		// Test 31 ensures when using operator '==' and a wildcard for the
		// dependency minor version the same version of dependency and component
		// matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 32 ensures when using operator '>=' the same version of dependency
		// and component matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 33 ensures when using operator '<=' and a wildcard for the
		// dependency minor version the same version of dependency and component
		// matches.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "<= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: true,
		},

		// Test 34 ensures when using operator '<' and a wildcard for the dependency
		// minor version the same version of dependency and component does not
		// match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 35 ensures when using operator '>' and a wildcard for the dependency
		// minor version the same version of dependency and component does not
		// match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 36 ensures when using operator '!=' and a wildcard for the
		// dependency minor version the same version of dependency and component
		// does not match.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "1.7.1",
			},
			ExpectedResult: false,
		},

		// Test 37 ensures when using operator '>=' and a wildcard for the
		// dependency minor version a higher component patch version does match with
		// the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "2.7.1",
			},
			ExpectedResult: true,
		},

		// Test 38 ensures when using operator '>' and a wildcard for the dependency
		// minor version a higher component patch version matchtes with the given
		// dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "2.7.1",
			},
			ExpectedResult: true,
		},

		// Test 39 ensures when using operator '<' and a wildcard for the dependency
		// minor version a higher component patch version matchtes with the given
		// dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "2.7.1",
			},
			ExpectedResult: false,
		},

		// Test 40 ensures when using operator '==' and a wildcard for the
		// dependency minor version a lower component patch version does not match
		// with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "2.7.1",
			},
			ExpectedResult: false,
		},

		// Test 41 ensures when using operator '!=' and a wildcard for the
		// dependency minor version a lower component patch version matchtes with
		// the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "2.7.1",
			},
			ExpectedResult: true,
		},

		// Test 42 ensures when using operator '>=' and a wildcard for the
		// dependency minor version a lower component patch version does not match
		// with the given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "0.7.1",
			},
			ExpectedResult: false,
		},

		// Test 43 ensures when using operator '>' and a wildcard for the dependency
		// minor version a lower component patch version does not match with the
		// given dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "0.7.1",
			},
			ExpectedResult: false,
		},

		// Test 44 ensures when using operator '<' and a wildcard for the dependency
		// minor version a lower component patch version matchtes with the given
		// dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "0.7.1",
			},
			ExpectedResult: true,
		},

		// Test 45 ensures when using operator '!=' and a wildcard for the dependency
		// minor version a lower component patch version matchtes with the given
		// dependency.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.x.x",
			},
			Component: Component{
				Name:    "kubernetes",
				Version: "0.7.1",
			},
			ExpectedResult: true,
		},
	}

	for i, tc := range testCases {
		b := tc.Dependency.Matches(tc.Component)
		if b != tc.ExpectedResult {
			t.Fatalf("test %d expected %#v got %#v", i, true, false)
		}
	}
}

func Test_Dependency_Validate(t *testing.T) {
	testCases := []struct {
		Dependency   Dependency
		ErrorMatcher func(err error) bool
	}{
		// Test 0 ensures an empty dependency is not valid.
		// and component matches.
		{
			Dependency:   Dependency{},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 1 is the same as 0 but with initialized properties.
		{
			Dependency: Dependency{
				Name:    "",
				Version: "",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 2 ensures a dependency missing a version is not valid.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 3 ensures a dependency missing a name is not valid.
		{
			Dependency: Dependency{
				Name:    "",
				Version: "== 1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 4 ensures a dependency missing a version operant is not valid.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 5 ensures a dependency using an invalid version operant is not
		// valid.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "$% 1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 6 is the same as 5 but with a different invalid version operant.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "^x 1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 7 is the same as 5 but with a different invalid version operant.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "^ 1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 8 is the same as 5 but with a different invalid version operant.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "x 1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 9 ensures a dependency not separating version operant by space is
		// invalid.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "==1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 10 is the same as 9 but with a different operant.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "<1.0.0",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 11 ensures a non-semver version is invalid.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== foo",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 12 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 13 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 14 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.x",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 15 is the same as 11 but with a different version.
		//
		// NOTE using the wildcard for minor does not allow a wildcard for patch.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.x.2",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 16 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.x.x.x",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 17 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.3.9.4",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 18 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.3.foo",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 19 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.bar.3",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 20 is the same as 11 but with a different version.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== kubernetes.14.3",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 21 is the same as 11 but with a different version.
		//
		// NOTE negative version numbers are not allowed.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.-3.7",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 22 is the same as 18 but with a different version.
		//
		// NOTE negative version numbers are not allowed.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== -1.3.7",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 23 is the same as 18 but with a different version.
		//
		// NOTE negative version numbers are not allowed.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.3.-7",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 24 ensures using the wildcard for the major version is not valid.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== x.x.x",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 25 is the same as 17 but with different input.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== x.3.9",
			},
			ErrorMatcher: IsInvalidDependency,
		},

		// Test 26 ensures a valid dependency does not throw an error.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 1.3.9",
			},
			ErrorMatcher: nil,
		},

		// Test 27 ensures a valid dependency does not throw an error.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 0.0.0",
			},
			ErrorMatcher: nil,
		},

		// Test 28 ensures a valid dependency does not throw an error.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "== 11.3.785",
			},
			ErrorMatcher: nil,
		},

		// Test 29 is the same as 25 but with a different operator.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "< 1.3.9",
			},
			ErrorMatcher: nil,
		},

		// Test 30 is the same as 25 but with a different operator.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "> 1.3.9",
			},
			ErrorMatcher: nil,
		},

		// Test 31 is the same as 25 but with a different operator.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "<= 1.3.9",
			},
			ErrorMatcher: nil,
		},

		// Test 32 is the same as 25 but with a different operator.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: ">= 1.3.9",
			},
			ErrorMatcher: nil,
		},

		// Test 33 is the same as 25 but with a different operator.
		{
			Dependency: Dependency{
				Name:    "kubernetes",
				Version: "!= 1.3.9",
			},
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		err := tc.Dependency.Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}
