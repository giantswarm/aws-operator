package versionbundle

import (
	"reflect"
	"testing"
	"time"
)

func Test_Release_Changelogs(t *testing.T) {
	testCases := []struct {
		Bundles            []Bundle
		ExpectedChangelogs []Changelog
		ErrorMatcher       func(err error) bool
	}{
		// Test 0 ensures creating a release with a nil slice of bundles throws
		// an error when creating a new release type.
		{
			Bundles:            nil,
			ExpectedChangelogs: nil,
			ErrorMatcher:       IsInvalidConfig,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:            []Bundle{},
			ExpectedChangelogs: nil,
			ErrorMatcher:       IsInvalidConfig,
		},

		// Test 2 ensures computing the release changelogs when having a list
		// of one bundle given works as expected.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "kubernetes",
							Description: "description",
							Kind:        "fixed",
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
					Version:    "0.0.1",
					WIP:        false,
				},
			},
			ExpectedChangelogs: []Changelog{
				{
					Component:   "kubernetes",
					Description: "description",
					Kind:        "fixed",
				},
			},
			ErrorMatcher: nil,
		},

		// Test 3 is the same as 2 but with a different changelogs.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "kubernetes",
							Description: "description",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "kube-dns",
							Version: "1.17.0",
						},
						{
							Name:    "calico",
							Version: "3.1.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: true,
					Name:       "kubernetes-operator",
					Time:       time.Unix(20, 15),
					Version:    "11.4.1",
					WIP:        false,
				},
			},
			ExpectedChangelogs: []Changelog{
				{
					Component:   "kubernetes",
					Description: "description",
					Kind:        "changed",
				},
			},
			ErrorMatcher: nil,
		},

		// Test 4 ensures computing the release changelogs when having a list of
		// two bundles given works as expected.
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
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			ExpectedChangelogs: []Changelog{
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
				{
					Component:   "etcd",
					Description: "Etcd version updated.",
					Kind:        "changed",
				},
				{
					Component:   "kubernetes",
					Description: "Kubernetes version updated.",
					Kind:        "changed",
				},
			},
			ErrorMatcher: nil,
		},

		// Test 5 is like 4 but with version bundles being flipped.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
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
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ExpectedChangelogs: []Changelog{
				{
					Component:   "etcd",
					Description: "Etcd version updated.",
					Kind:        "changed",
				},
				{
					Component:   "kubernetes",
					Description: "Kubernetes version updated.",
					Kind:        "changed",
				},
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
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		config := DefaultReleaseConfig()

		config.Bundles = tc.Bundles

		r, err := NewRelease(config)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		c := r.Changelogs()
		if !reflect.DeepEqual(c, tc.ExpectedChangelogs) {
			t.Fatalf("test %d expected %#v got %#v", i, tc.ExpectedChangelogs, c)
		}
	}
}

func Test_Release_Components(t *testing.T) {
	testCases := []struct {
		Bundles            []Bundle
		ExpectedComponents []Component
		ErrorMatcher       func(err error) bool
	}{
		// Test 0 ensures creating a release with a nil slice of bundles throws
		// an error when creating a new release type.
		{
			Bundles:            nil,
			ExpectedComponents: nil,
			ErrorMatcher:       IsInvalidConfig,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:            []Bundle{},
			ExpectedComponents: nil,
			ErrorMatcher:       IsInvalidConfig,
		},

		// Test 2 ensures computing the release components when having a list
		// of one bundle given works as expected.
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
					Version:    "0.0.1",
					WIP:        false,
				},
			},
			ExpectedComponents: []Component{
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
			},
			ErrorMatcher: nil,
		},

		// Test 3 is the same as 2 but with a different components.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "kube-dns",
							Version: "1.17.0",
						},
						{
							Name:    "calico",
							Version: "3.1.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: true,
					Name:       "kubernetes-operator",
					Time:       time.Unix(20, 15),
					Version:    "11.4.1",
					WIP:        false,
				},
			},
			ExpectedComponents: []Component{
				{
					Name:    "kube-dns",
					Version: "1.17.0",
				},
				{
					Name:    "calico",
					Version: "3.1.0",
				},
			},
			ErrorMatcher: nil,
		},

		// Test 4 ensures computing the release components when having a list of
		// two bundles given works as expected.
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
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			ExpectedComponents: []Component{
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
				{
					Name:    "etcd",
					Version: "3.2.0",
				},
				{
					Name:    "kubernetes",
					Version: "1.7.1",
				},
			},
			ErrorMatcher: nil,
		},

		// Test 5 is like 4 but with version bundles being flipped.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
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
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ExpectedComponents: []Component{
				{
					Name:    "etcd",
					Version: "3.2.0",
				},
				{
					Name:    "kubernetes",
					Version: "1.7.1",
				},
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
			},
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		config := DefaultReleaseConfig()

		config.Bundles = tc.Bundles

		r, err := NewRelease(config)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		c := r.Components()
		if !reflect.DeepEqual(c, tc.ExpectedComponents) {
			t.Fatalf("test %d expected %#v got %#v", i, tc.ExpectedComponents, c)
		}
	}
}

func Test_Release_Deprecated(t *testing.T) {
	testCases := []struct {
		Bundles            []Bundle
		ExpectedDeprecated bool
		ErrorMatcher       func(err error) bool
	}{
		// Test 0 ensures creating a release with a nil slice of bundles throws
		// an error when creating a new release type.
		{
			Bundles:            nil,
			ExpectedDeprecated: false,
			ErrorMatcher:       IsInvalidConfig,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:            []Bundle{},
			ExpectedDeprecated: false,
			ErrorMatcher:       IsInvalidConfig,
		},

		// Test 2 ensures computing the release deprecated flag when having a list
		// of one bundle given works as expected.
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
					Version:    "0.0.1",
					WIP:        false,
				},
			},
			ExpectedDeprecated: false,
			ErrorMatcher:       nil,
		},

		// Test 3 is the same as 2 but with a different deprecated flag.
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
					Deprecated: true,
					Name:       "kubernetes-operator",
					Time:       time.Unix(20, 15),
					Version:    "11.4.1",
					WIP:        false,
				},
			},
			ExpectedDeprecated: true,
			ErrorMatcher:       nil,
		},

		// Test 4 ensures computing the release deprecated flag when having a list of
		// two bundles given works as expected.
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
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			ExpectedDeprecated: false,
			ErrorMatcher:       nil,
		},

		// Test 5 is like 4 but with version bundles being flipped.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
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
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ExpectedDeprecated: false,
			ErrorMatcher:       nil,
		},

		// Test 6 is like 4 but with all deprecated flags being true.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   true,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
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
					Deprecated: true,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ExpectedDeprecated: true,
			ErrorMatcher:       nil,
		},

		// Test 7 is like 4 but with only one deprecated flag being true.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
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
					Deprecated: true,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ExpectedDeprecated: true,
			ErrorMatcher:       nil,
		},

		// Test 8 is like 7 but with version bundles being flipped.
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
					Deprecated: true,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			ExpectedDeprecated: true,
			ErrorMatcher:       nil,
		},
	}

	for i, tc := range testCases {
		config := DefaultReleaseConfig()

		config.Bundles = tc.Bundles

		r, err := NewRelease(config)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		d := r.Deprecated()
		if d != tc.ExpectedDeprecated {
			t.Fatalf("test %d expected %t got %t", i, tc.ExpectedDeprecated, d)
		}
	}
}

func Test_Release_Timestamp(t *testing.T) {
	testCases := []struct {
		Bundles           []Bundle
		ExpectedTimestamp string
		ErrorMatcher      func(err error) bool
	}{
		// Test 0 ensures creating a release with a nil slice of bundles throws
		// an error when creating a new release type.
		{
			Bundles:           nil,
			ExpectedTimestamp: "",
			ErrorMatcher:      IsInvalidConfig,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:           []Bundle{},
			ExpectedTimestamp: "",
			ErrorMatcher:      IsInvalidConfig,
		},

		// Test 2 ensures computing the release timestamp when having a list of one
		// bundle given works as expected.
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
					Time:       time.Unix(10, 5).In(time.UTC),
					Version:    "0.0.1",
					WIP:        false,
				},
			},
			ExpectedTimestamp: "1970-01-01T00:00:10.000000Z",
			ErrorMatcher:      nil,
		},

		// Test 3 is the same as 2 but with a different timestamp.
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
					Time:       time.Unix(20, 15).In(time.UTC),
					Version:    "11.4.1",
					WIP:        false,
				},
			},
			ExpectedTimestamp: "1970-01-01T00:00:20.000000Z",
			ErrorMatcher:      nil,
		},

		// Test 4 ensures computing the release timestamp when having a list of
		// two bundles given works as expected.
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
					Time:       time.Unix(10, 5).In(time.UTC),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15).In(time.UTC),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			ExpectedTimestamp: "1970-01-01T00:00:20.000000Z",
			ErrorMatcher:      nil,
		},

		// Test 5 is like 4 but with version bundles being flipped.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15).In(time.UTC),
					Version:      "0.2.0",
					WIP:          false,
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
					Time:       time.Unix(10, 5).In(time.UTC),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ExpectedTimestamp: "1970-01-01T00:00:20.000000Z",
			ErrorMatcher:      nil,
		},
	}

	for i, tc := range testCases {
		config := DefaultReleaseConfig()

		config.Bundles = tc.Bundles

		r, err := NewRelease(config)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		ts := r.Timestamp()
		if ts != tc.ExpectedTimestamp {
			t.Fatalf("test %d expected %s got %s", i, tc.ExpectedTimestamp, ts)
		}
	}
}

func Test_Release_Version(t *testing.T) {
	testCases := []struct {
		Bundles         []Bundle
		ExpectedVersion string
		ErrorMatcher    func(err error) bool
	}{
		// Test 0 ensures creating a release with a nil slice of bundles throws
		// an error when creating a new release type.
		{
			Bundles:         nil,
			ExpectedVersion: "",
			ErrorMatcher:    IsInvalidConfig,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:         []Bundle{},
			ExpectedVersion: "",
			ErrorMatcher:    IsInvalidConfig,
		},

		// Test 2 ensures computing the release version when having a list of
		// one bundle given works as expected.
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
					Version:    "0.0.1",
					WIP:        false,
				},
			},
			ExpectedVersion: "0.0.1",
			ErrorMatcher:    nil,
		},

		// Test 3 is the same as 2 but with a different version.
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
					Version:    "11.4.1",
					WIP:        false,
				},
			},
			ExpectedVersion: "11.4.1",
			ErrorMatcher:    nil,
		},

		// Test 4 ensures computing the release version when having a list of
		// two bundles given works as expected.
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
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "0.2.0",
					WIP:          false,
				},
			},
			ExpectedVersion: "0.3.0",
			ErrorMatcher:    nil,
		},

		// Test 5 is the same as 4 but with a different version.
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
					Version:    "5.0.1",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.2.0",
						},
						{
							Name:    "kubernetes",
							Version: "1.7.1",
						},
					},
					Dependencies: []Dependency{},
					Name:         "cloud-config-operator",
					Deprecated:   false,
					Time:         time.Unix(20, 15),
					Version:      "12.2.77",
					WIP:          false,
				},
			},
			ExpectedVersion: "17.2.78",
			ErrorMatcher:    nil,
		},
	}

	for i, tc := range testCases {
		config := DefaultReleaseConfig()

		config.Bundles = tc.Bundles

		r, err := NewRelease(config)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		v := r.Version()
		if v != tc.ExpectedVersion {
			t.Fatalf("test %d expected %s got %s", i, tc.ExpectedVersion, v)
		}
	}
}
