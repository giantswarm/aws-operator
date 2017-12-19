package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterRecordSetsRegularFields(t *testing.T) {
	testCases := []struct {
		description                   string
		customObject                  v1alpha1.AWSConfig
		expectedAPIHostedZone         string
		expectedAPIDomain             string
		expectedEtcdHostedZone        string
		expectedEtcdDomain            string
		expectedIngressHostedZone     string
		expectedIngressDomain         string
		expectedIngressWildcardDomain string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								Domain: "api.domain",
							},
							IngressController: v1alpha1.ClusterKubernetesIngressController{
								Domain:         "ingress.domain",
								WildcardDomain: "ingressWildcardDomain",
							},
						},
						Etcd: v1alpha1.ClusterEtcd{
							Domain: "etcd.domain",
						},
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						API: v1alpha1.AWSConfigSpecAWSAPI{
							HostedZones: "apiHostedZones",
						},
						Etcd: v1alpha1.AWSConfigSpecAWSEtcd{
							HostedZones: "etcdHostedZone",
						},
						Ingress: v1alpha1.AWSConfigSpecAWSIngress{
							HostedZones: "ingressHostedZone",
						},
					},
				},
			},
			expectedAPIHostedZone:         "apiHostedZones",
			expectedAPIDomain:             "api.domain",
			expectedEtcdHostedZone:        "etcdHostedZone",
			expectedEtcdDomain:            "etcd.domain",
			expectedIngressHostedZone:     "ingressHostedZone",
			expectedIngressDomain:         "ingress.domain",
			expectedIngressWildcardDomain: "ingressWildcardDomain",
		},
	}

	clients := Clients{
		EC2: &EC2ClientMock{},
		ELB: &ELBClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			err := a.getRecordSets(tc.customObject, clients)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.APIELBHostedZones != tc.expectedAPIHostedZone {
				t.Errorf("unexpected APIELBHostedZones, got %q, want %q", a.APIELBHostedZones, tc.expectedAPIHostedZone)
			}
			if a.APIELBDomain != tc.expectedAPIDomain {
				t.Errorf("unexpected APIELBDomain, got %q, want %q", a.APIELBDomain, tc.expectedAPIDomain)
			}
			if a.EtcdELBHostedZones != tc.expectedEtcdHostedZone {
				t.Errorf("unexpected EtcdELBHostedZones, got %q, want %q", a.EtcdELBHostedZones, tc.expectedEtcdHostedZone)
			}
			if a.EtcdELBDomain != tc.expectedEtcdDomain {
				t.Errorf("unexpected EtcdELBDomain, got %q, want %q", a.EtcdELBDomain, tc.expectedEtcdDomain)
			}
			if a.IngressELBHostedZones != tc.expectedIngressHostedZone {
				t.Errorf("unexpected IngressELBHostedZones, got %q, want %q", a.IngressELBHostedZones, tc.expectedIngressHostedZone)
			}
			if a.IngressELBDomain != tc.expectedIngressDomain {
				t.Errorf("unexpected IngressELBDomain, got %q, want %q", a.IngressELBDomain, tc.expectedIngressDomain)
			}
			if a.IngressWildcardELBDomain != tc.expectedIngressWildcardDomain {
				t.Errorf("unexpected IngressWildcardELBDomain, got %q, want %q", a.IngressWildcardELBDomain, tc.expectedIngressWildcardDomain)
			}

		})
	}
}

func TestAdapterRecordSetsAPIDNS(t *testing.T) {
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain: "api.domain",
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain: "ingress.domain",
					},
				},
				Etcd: v1alpha1.ClusterEtcd{
					Domain: "etcd.domain",
				},
			},
		},
	}
	testCases := []struct {
		description string
		lbName      string
		expectedDNS string
	}{
		{
			description: "basic match",
			lbName:      "test-cluster-api",
			expectedDNS: "myDNS",
		},
	}

	clients := Clients{
		EC2: &EC2ClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			clients.ELB = &ELBClientMock{
				dns:  tc.expectedDNS,
				name: tc.lbName,
			}
			err := a.getRecordSets(customObject, clients)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if a.APIELBDNS != tc.expectedDNS {
				t.Fatalf("unexpected APIELBDNS, got %q, want %q", a.APIELBDNS, tc.expectedDNS)
			}
		})
	}
}

func TestAdapterRecordSetsIngressDNS(t *testing.T) {
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain: "api.domain",
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain: "ingress.domain",
					},
				},
				Etcd: v1alpha1.ClusterEtcd{
					Domain: "etcd.domain",
				},
			},
		},
	}
	testCases := []struct {
		description string
		lbName      string
		expectedDNS string
	}{
		{
			description: "basic match",
			lbName:      "test-cluster-ingress",
			expectedDNS: "myDNS",
		},
	}

	clients := Clients{
		EC2: &EC2ClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			clients.ELB = &ELBClientMock{
				dns:  tc.expectedDNS,
				name: tc.lbName,
			}
			err := a.getRecordSets(customObject, clients)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if a.IngressELBDNS != tc.expectedDNS {
				t.Fatalf("unexpected IngressELBDNS, got %q, want %q", a.IngressELBDNS, tc.expectedDNS)
			}
		})
	}
}

func TestAdapterRecordSetsEtcdDNS(t *testing.T) {
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain: "api.domain",
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain: "ingress.domain",
					},
				},
				Etcd: v1alpha1.ClusterEtcd{
					Domain: "etcd.domain",
				},
			},
		},
	}
	testCases := []struct {
		description string
		lbName      string
		expectedDNS string
	}{
		{
			description: "basic match",
			lbName:      "test-cluster-etcd",
			expectedDNS: "myDNS",
		},
	}

	clients := Clients{
		EC2: &EC2ClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			clients.ELB = &ELBClientMock{
				dns:  tc.expectedDNS,
				name: tc.lbName,
			}
			err := a.getRecordSets(customObject, clients)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if a.EtcdELBDNS != tc.expectedDNS {
				t.Fatalf("unexpected EtcdELBDNS, got %q, want %q", a.EtcdELBDNS, tc.expectedDNS)
			}
		})
	}
}
