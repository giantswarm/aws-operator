package versionbundle

import (
	"testing"
	"time"
)

func Test_AggregatedBundles_Validate(t *testing.T) {
	testCases := []struct {
		AggregatedBundles [][]Bundle
		ErrorMatcher      func(err error) bool
	}{
		// Test 0 ensures that validating an aggregated list of version bundles
		// having different lengths throws an error.
		{
			AggregatedBundles: [][]Bundle{
				{
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
				{
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
						Deprecated:   false,
						Name:         "cloud-config-operator",
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
			},
			ErrorMatcher: IsInvalidAggregatedBundlesError,
		},

		// Test 1 ensures that validating an aggregated list of version bundles
		// having the same version bundles throws an error.
		{
			AggregatedBundles: [][]Bundle{
				{
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
						Deprecated:   false,
						Name:         "cloud-config-operator",
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
				{
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
						Deprecated:   false,
						Name:         "cloud-config-operator",
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
			},
			ErrorMatcher: IsInvalidAggregatedBundlesError,
		},
	}

	for i, tc := range testCases {
		err := AggregatedBundles(tc.AggregatedBundles).Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}
