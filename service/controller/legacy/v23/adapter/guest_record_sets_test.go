package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterRecordSetsRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description            string
		customObject           v1alpha1.AWSConfig
		route53Enabled         bool
		expectedBaseDomain     string
		expectedClusterID      string
		expectedRoute53Enabled bool
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
						HostedZones: v1alpha1.AWSConfigSpecAWSHostedZones{
							API: v1alpha1.AWSConfigSpecAWSHostedZonesZone{
								Name: "installation.aws.eu-central-1.gigantic.io",
							},
						},
					},
				},
			},
			route53Enabled:         true,
			expectedRoute53Enabled: true,
			expectedClusterID:      "test-cluster",
			expectedBaseDomain:     "installation.aws.eu-central-1.gigantic.io",
		},
	}

	clients := Clients{}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject:   tc.customObject,
				Clients:        clients,
				Route53Enabled: tc.route53Enabled,
			}
			err := a.Guest.RecordSets.Adapt(cfg)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.RecordSets.BaseDomain != tc.expectedBaseDomain {
				t.Fatalf("BaseDomain == %q, want %q", a.Guest.RecordSets.BaseDomain, tc.expectedBaseDomain)
			}
			if a.Guest.RecordSets.ClusterID != tc.expectedClusterID {
				t.Fatalf("ClusterID == %q, want %q", a.Guest.RecordSets.ClusterID, tc.expectedClusterID)
			}
			if a.Guest.RecordSets.Route53Enabled != tc.expectedRoute53Enabled {
				t.Fatalf("Route53Enabled == %v, want %v", a.Guest.RecordSets.Route53Enabled, tc.expectedRoute53Enabled)
			}
		})
	}
}
