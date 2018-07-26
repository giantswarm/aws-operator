package migration

import (
	"context"
	"reflect"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_migrateSpec(t *testing.T) {
	testCases := []struct {
		name         string
		spec         providerv1alpha1.AWSConfigSpec
		expectedSpec providerv1alpha1.AWSConfigSpec
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0: fill missing fields",
			spec: providerv1alpha1.AWSConfigSpec{
				Cluster: providerv1alpha1.Cluster{
					Kubernetes: providerv1alpha1.ClusterKubernetes{
						API: providerv1alpha1.ClusterKubernetesAPI{
							Domain: "api.eggs2.k8s.gauss.eu-central-1.aws.gigantic.io",
						},
					},
				},
			},
			expectedSpec: providerv1alpha1.AWSConfigSpec{
				AWS: providerv1alpha1.AWSConfigSpecAWS{
					CredentialSecret: providerv1alpha1.CredentialSecret{
						Name:      "credential-default",
						Namespace: "giantswarm",
					},
					HostedZones: providerv1alpha1.AWSConfigSpecAWSHostedZones{
						API: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "gauss.eu-central-1.aws.gigantic.io",
						},
						Etcd: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "gauss.eu-central-1.aws.gigantic.io",
						},
						Ingress: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "gauss.eu-central-1.aws.gigantic.io",
						},
					},
				},
				Cluster: providerv1alpha1.Cluster{
					Kubernetes: providerv1alpha1.ClusterKubernetes{
						API: providerv1alpha1.ClusterKubernetesAPI{
							Domain: "api.eggs2.k8s.gauss.eu-central-1.aws.gigantic.io",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: not mess with fields already set",
			spec: providerv1alpha1.AWSConfigSpec{
				AWS: providerv1alpha1.AWSConfigSpecAWS{
					CredentialSecret: providerv1alpha1.CredentialSecret{
						Name:      "test-credential",
						Namespace: "test-credential-namespace",
					},
					HostedZones: providerv1alpha1.AWSConfigSpecAWSHostedZones{
						API: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "test-api.gigantic.io",
						},
						Etcd: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "test-etcd.gigantic.io",
						},
						Ingress: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "test-ingress.gigantic.io",
						},
					},
				},
				Cluster: providerv1alpha1.Cluster{
					Kubernetes: providerv1alpha1.ClusterKubernetes{
						API: providerv1alpha1.ClusterKubernetesAPI{
							Domain: "api.eggs5.k8s.gauss.eu-central-1.aws.gigantic.io",
						},
					},
				},
			},
			expectedSpec: providerv1alpha1.AWSConfigSpec{
				AWS: providerv1alpha1.AWSConfigSpecAWS{
					CredentialSecret: providerv1alpha1.CredentialSecret{
						Name:      "test-credential",
						Namespace: "test-credential-namespace",
					},
					HostedZones: providerv1alpha1.AWSConfigSpecAWSHostedZones{
						API: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "test-api.gigantic.io",
						},
						Etcd: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "test-etcd.gigantic.io",
						},
						Ingress: providerv1alpha1.AWSConfigSpecAWSHostedZonesZone{
							Name: "test-ingress.gigantic.io",
						},
					},
				},
				Cluster: providerv1alpha1.Cluster{
					Kubernetes: providerv1alpha1.ClusterKubernetes{
						API: providerv1alpha1.ClusterKubernetesAPI{
							Domain: "api.eggs5.k8s.gauss.eu-central-1.aws.gigantic.io",
						},
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Resource{
				logger: microloggertest.New(),
			}

			err := r.migrateSpec(context.Background(), &tc.spec)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if tc.errorMatcher != nil {
				return
			}

			if !reflect.DeepEqual(tc.spec, tc.expectedSpec) {
				t.Errorf("spec == %q, want %q", tc.spec, tc.expectedSpec)
			}
		})
	}
}

func Test_zoneFromAPIDomain(t *testing.T) {
	testCases := []struct {
		name         string
		apiDomain    string
		expectedZone string
		errorMatcher func(err error) bool
	}{
		{
			name:         "case 0: normal case",
			apiDomain:    "api.eggs2.k8s.gauss.eu-central-1.aws.gigantic.io",
			expectedZone: "gauss.eu-central-1.aws.gigantic.io",
		},
		{
			name:         "case 1: domain too short",
			apiDomain:    "api.eggs2.k8s.gigantic",
			errorMatcher: IsMalformedDomain,
		},
		{
			name:         "case 1: minimal length domain",
			apiDomain:    "api.eggs2.k8s.gigantic.io",
			expectedZone: "gigantic.io",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			zone, err := zoneFromAPIDomain(tc.apiDomain)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if tc.errorMatcher != nil {
				return
			}

			if zone != tc.expectedZone {
				t.Fatalf("zone == %q, want %q", zone, tc.expectedZone)
			}
		})
	}
}
