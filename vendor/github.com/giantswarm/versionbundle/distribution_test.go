package versionbundle

import (
	"testing"
	"time"
)

func Test_Distribution_Version(t *testing.T) {
	testCases := []struct {
		Bundles         []Bundle
		ExpectedVersion string
		ErrorMatcher    func(err error) bool
	}{
		// Test 0 ensures creating a distribution with a nil slice of bundles throws
		// an error when creating a new distribution type.
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

		// Test 2 ensures computing the distribution version when having a list of
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

		// Test 4 ensures computing the distribution version when having a list of
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
		config := DefaultDistributionConfig()

		config.Bundles = tc.Bundles

		d, err := NewDistribution(config)
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		v := d.Version()
		if v != tc.ExpectedVersion {
			t.Fatalf("test %d expected %s got %s", i, tc.ExpectedVersion, v)
		}
	}
}
