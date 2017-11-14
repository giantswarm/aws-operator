package versionbundle

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func Test_Bundles_Contains(t *testing.T) {
	testCases := []struct {
		Bundles        []Bundle
		Bundle         Bundle
		ExpectedResult bool
	}{
		// Test 0 ensures that a nil list and an empty bundle result in false.
		{
			Bundles:        nil,
			Bundle:         Bundle{},
			ExpectedResult: false,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:        []Bundle{},
			Bundle:         Bundle{},
			ExpectedResult: false,
		},

		// Test 2 ensures a list containing one version bundle and a matching
		// version bundle results in true.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "kubernetes-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.1.0",
				WIP:          false,
			},
			ExpectedResult: true,
		},

		// Test 3 ensures a list containing two version bundle and a matching
		// version bundle results in true.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.2.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.2.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "kubernetes-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.2.0",
				WIP:          false,
			},
			ExpectedResult: true,
		},

		// Test 4 ensures a list containing one version bundle and a version bundle
		// that does not match results in false.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.2.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "kubernetes-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.2.0",
				WIP:          false,
			},
			ExpectedResult: false,
		},

		// Test 5 ensures a list containing two version bundle and a version bundle
		// that does not match results in false.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.2.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.3.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "kubernetes-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.3.0",
				WIP:          false,
			},
			ExpectedResult: false,
		},
	}

	for i, tc := range testCases {
		result := Bundles(tc.Bundles).Contain(tc.Bundle)
		if result != tc.ExpectedResult {
			t.Fatalf("test %d expected %#v got %#v", i, tc.ExpectedResult, result)
		}
	}
}

func Test_Bundles_Copy(t *testing.T) {
	bundles := []Bundle{
		{
			Changelogs: []Changelog{},
			Components: []Component{
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
			},
			Dependencies: []Dependency{
				{
					Name:    "kubernetes",
					Version: "<= 1.7.x",
				},
			},
			Deprecated: false,
			Name:       "kubernetes-operator",
			Time:       time.Unix(10, 5),
			Version:    "0.1.0",
			WIP:        false,
		},
		{
			Changelogs: []Changelog{
				{
					Component:   "calico",
					Description: "Calico version updated.",
					Kind:        "changed",
				},
				{
					Component:   "kubernetes",
					Description: "Kubernetes version requirements changed due to calico update.",
					Kind:        "changed",
				},
			},
			Components: []Component{
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
			},
			Dependencies: []Dependency{
				{
					Name:    "kubernetes",
					Version: "<= 1.7.x",
				},
			},
			Deprecated: false,
			Name:       "kubernetes-operator",
			Time:       time.Unix(20, 10),
			Version:    "0.0.9",
			WIP:        false,
		},
	}

	b1 := CopyBundles(bundles)
	b2 := CopyBundles(bundles)

	sort.Sort(SortBundlesByTime(b1))
	sort.Sort(SortBundlesByVersion(b2))

	if reflect.DeepEqual(b1, b2) {
		t.Fatalf("expected %#v got %#v", b1, b2)
	}
}

func Test_Bundles_GetBundleByName(t *testing.T) {
	testCases := []struct {
		Bundles        []Bundle
		Name           string
		ExpectedBundle Bundle
		ErrorMatcher   func(err error) bool
	}{
		// Test 0 ensures that a nil list and an empty name throws an execution
		// failed error.
		{
			Bundles:        nil,
			Name:           "",
			ExpectedBundle: Bundle{},
			ErrorMatcher:   IsExecutionFailed,
		},

		// Test 1 ensures that a nil list and a non-empty name throws an execution
		// failed error.
		{
			Bundles:        nil,
			Name:           "kubernetes-operator",
			ExpectedBundle: Bundle{},
			ErrorMatcher:   IsExecutionFailed,
		},

		// Test 2 ensures that a non-empty list and an empty name throws an execution
		// failed error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			Name:           "",
			ExpectedBundle: Bundle{},
			ErrorMatcher:   IsExecutionFailed,
		},

		// Test 3 ensures that a non-empty list and an non-empty name throws a
		// not found errorn case the given name does not exist in the given list.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			Name:           "cert-operator",
			ExpectedBundle: Bundle{},
			ErrorMatcher:   IsBundleNotFound,
		},

		// Test 4 is the same as 3 but with different version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.5",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "cloud-config-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			Name:           "cert-operator",
			ExpectedBundle: Bundle{},
			ErrorMatcher:   IsBundleNotFound,
		},

		// Test 5 ensures that a bundle can be found.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.5",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "cloud-config-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			Name: "cloud-config-operator",
			ExpectedBundle: Bundle{
				Changelogs: []Changelog{},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kubernetes",
						Version: "1.7.5",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "cloud-config-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.1.0",
				WIP:          false,
			},
			ErrorMatcher: nil,
		},

		// Test 6 is the same as 5 but with different bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.5",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "cloud-config-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			Name: "cloud-config-operator",
			ExpectedBundle: Bundle{
				Changelogs: []Changelog{},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kubernetes",
						Version: "1.7.5",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "cloud-config-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.1.0",
				WIP:          false,
			},
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		result, err := GetBundleByName(tc.Bundles, tc.Name)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		} else {
			if !reflect.DeepEqual(result, tc.ExpectedBundle) {
				t.Fatalf("test %d expected %#v got %#v", i, tc.ExpectedBundle, result)
			}
		}
	}
}

func Test_Bundles_Validate(t *testing.T) {
	testCases := []struct {
		Bundles      []Bundle
		ErrorMatcher func(err error) bool
	}{
		// Test 0 ensures that a nil list is not valid.
		{
			Bundles:      nil,
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:      []Bundle{},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 2 ensures validation of a list of version bundles where any version
		// bundle has no changelogs throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 3 is the same as 2 but with multiple version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 4 ensures validation of a list of version bundles where any version
		// bundle has no components throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 5 is the same as 4 but with multiple version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 6 ensures validation of a list of version bundles where any version
		// bundle has no dependency does not throw an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			ErrorMatcher: nil,
		},

		// Test 7 is the same as 6 but with multiple version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: nil,
		},

		// Test 8 ensures validation of a list of version bundles not having the
		// same name throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "ingress-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 9 ensures validation of a list of version bundles having duplicated
		// version bundles throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 10 ensures validation of a list of version bundles having the same
		// version throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "kube-dns",
							Description: "Kube-DNS version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.1.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(20, 10),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 11 ensures validation of a list of version bundles in which a newer
		// version bundle (time) has a lower version number throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "kube-dns",
							Description: "Kube-DNS version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.1.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(20, 10),
					Version:      "0.0.9",
					WIP:          false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},
	}

	for i, tc := range testCases {
		err := Bundles(tc.Bundles).Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}
