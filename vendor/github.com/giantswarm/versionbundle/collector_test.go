package versionbundle

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/go-resty/resty"
)

func Test_Collector_Collect(t *testing.T) {
	testCases := []struct {
		HandlerFuncs    []func(w http.ResponseWriter, r *http.Request)
		ExpectedBundles []Bundle
	}{
		// Test 0 ensures a single version bundle exposed by a single endpoint
		// results in a single version bundle.
		{
			HandlerFuncs: []func(w http.ResponseWriter, r *http.Request){
				func(w http.ResponseWriter, r *http.Request) {
					cr := CollectorEndpointResponse{
						VersionBundles: []Bundle{
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
								Time:         time.Unix(10, 5).In(time.UTC),
								Version:      "0.1.0",
								WIP:          false,
							},
						},
					}
					b, err := json.Marshal(cr)
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
					_, err = io.WriteString(w, string(b))
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
				},
			},
			ExpectedBundles: []Bundle{
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
					Time:         time.Unix(10, 5).In(time.UTC),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
		},

		// Test 1 ensures one version bundle exposed by a one endpoint and another
		// version bundle exposed by another endpoint results in two version
		// bundles.
		{
			HandlerFuncs: []func(w http.ResponseWriter, r *http.Request){
				func(w http.ResponseWriter, r *http.Request) {
					cr := CollectorEndpointResponse{
						VersionBundles: []Bundle{
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
								Time:         time.Unix(10, 5).In(time.UTC),
								Version:      "0.1.0",
								WIP:          false,
							},
						},
					}
					b, err := json.Marshal(cr)
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
					_, err = io.WriteString(w, string(b))
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
				},
				func(w http.ResponseWriter, r *http.Request) {
					cr := CollectorEndpointResponse{
						VersionBundles: []Bundle{
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
					}
					b, err := json.Marshal(cr)
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
					_, err = io.WriteString(w, string(b))
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
				},
			},
			ExpectedBundles: []Bundle{
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
					Time:         time.Unix(10, 5).In(time.UTC),
					Version:      "0.1.0",
					WIP:          false,
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
		},

		// Test 2 ensures two version bundles exposed by a one endpoint and another
		// two version bundles exposed by another endpoint results in four version
		// bundles.
		{
			HandlerFuncs: []func(w http.ResponseWriter, r *http.Request){
				func(w http.ResponseWriter, r *http.Request) {
					cr := CollectorEndpointResponse{
						VersionBundles: []Bundle{
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
								Time:         time.Unix(10, 5).In(time.UTC),
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
								Time:         time.Unix(10, 5).In(time.UTC),
								Version:      "0.2.0",
								WIP:          false,
							},
						},
					}
					b, err := json.Marshal(cr)
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
					_, err = io.WriteString(w, string(b))
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
				},
				func(w http.ResponseWriter, r *http.Request) {
					cr := CollectorEndpointResponse{
						VersionBundles: []Bundle{
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
										Component:   "etcd",
										Description: "Etcd version updated.",
										Kind:        "changed",
									},
								},
								Components: []Component{
									{
										Name:    "etcd",
										Version: "3.3.0",
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
								Version:      "0.3.0",
								WIP:          false,
							},
						},
					}
					b, err := json.Marshal(cr)
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
					_, err = io.WriteString(w, string(b))
					if err != nil {
						t.Fatalf("expected %#v got %#v", nil, err)
					}
				},
			},
			ExpectedBundles: []Bundle{
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
					Time:         time.Unix(10, 5).In(time.UTC),
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
					Time:         time.Unix(10, 5).In(time.UTC),
					Version:      "0.2.0",
					WIP:          false,
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
				{
					Changelogs: []Changelog{
						{
							Component:   "etcd",
							Description: "Etcd version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "etcd",
							Version: "3.3.0",
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
					Version:      "0.3.0",
					WIP:          false,
				},
			},
		},
	}

	for i, tc := range testCases {
		var endpoints []*url.URL
		for _, hf := range tc.HandlerFuncs {
			ts := httptest.NewServer(http.HandlerFunc(hf))
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("test %d expected %#v got %#v", i, nil, err)
			}
			endpoints = append(endpoints, u)
		}

		var err error

		var collector *Collector
		{
			c := DefaultCollectorConfig()

			c.RestClient = resty.New()

			c.Endpoints = endpoints

			collector, err = NewCollector(c)
			if err != nil {
				t.Fatalf("test %d expected %#v got %#v", i, nil, err)
			}
		}

		b1 := collector.Bundles()
		if b1 != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, b1)
		}

		err = collector.Collect(context.TODO())
		if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}

		b2 := collector.Bundles()
		if !reflect.DeepEqual(b2, tc.ExpectedBundles) {
			t.Fatalf("test %d expected %#v got %#v", i, tc.ExpectedBundles, b2)
		}
	}
}
