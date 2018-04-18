package adapter

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v3/key"
)

func TestAdapterLoadBalancersRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description                              string
		customObject                             v1alpha1.AWSConfig
		errorMatcher                             func(error) bool
		expectedAPIElbIdleTimoutSeconds          int
		expectedAPIElbName                       string
		expectedAPIElbPortsToOpen                portPairs
		expectedAPIElbScheme                     string
		expectedAPIElbSecurityGroupID            string
		expectedAPIElbSubnetID                   string
		expectedELBAZ                            string
		expectedELBHealthCheckHealthyThreshold   int
		expectedELBHealthCheckInterval           int
		expectedELBHealthCheckTimeout            int
		expectedELBHealthCheckUnhealthyThreshold int
		expectedIngressElbIdleTimoutSeconds      int
		expectedIngressElbName                   string
		expectedIngressElbPortsToOpen            portPairs
		expectedIngressElbScheme                 string
	}{
		{
			description:  "empty custom object",
			customObject: v1alpha1.AWSConfig{},
			errorMatcher: key.IsMissingCloudConfigKey,
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								Domain:     "api.test-cluster.aws.giantswarm.io",
								SecurePort: 443,
							},
							IngressController: v1alpha1.ClusterKubernetesIngressController{
								Domain:       "ingress.test-cluster.aws.giantswarm.io",
								InsecurePort: 30010,
								SecurePort:   30011,
							},
						},
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						API: v1alpha1.AWSConfigSpecAWSAPI{
							ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
								IdleTimeoutSeconds: 3600,
							},
						},
						AZ: "eu-central-1a",
						Ingress: v1alpha1.AWSConfigSpecAWSIngress{
							ELB: v1alpha1.AWSConfigSpecAWSIngressELB{
								IdleTimeoutSeconds: 60,
							},
						},
					},
				},
			},
			errorMatcher:                    nil,
			expectedAPIElbIdleTimoutSeconds: 3600,
			expectedAPIElbName:              "test-cluster-api",
			expectedAPIElbPortsToOpen: portPairs{
				{
					PortELB:      443,
					PortInstance: 443,
				},
			},
			expectedAPIElbScheme:                     "internet-facing",
			expectedELBAZ:                            "eu-central-1a",
			expectedELBHealthCheckHealthyThreshold:   2,
			expectedELBHealthCheckInterval:           5,
			expectedELBHealthCheckTimeout:            3,
			expectedELBHealthCheckUnhealthyThreshold: 2,
			expectedIngressElbIdleTimoutSeconds:      60,
			expectedIngressElbName:                   "test-cluster-ingress",
			expectedIngressElbPortsToOpen: portPairs{
				{
					PortELB:      443,
					PortInstance: 30011,
				},
				{
					PortELB:      80,
					PortInstance: 30010,
				},
			},
			expectedIngressElbScheme: "internet-facing",
		},
	}

	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getLoadBalancers(cfg)

			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if tc.expectedAPIElbIdleTimoutSeconds != a.APIElbIdleTimoutSeconds {
				t.Errorf("expected API ELB Idle Timeout Seconds, got %q, want %q", a.APIElbIdleTimoutSeconds, tc.expectedAPIElbIdleTimoutSeconds)
			}

			if tc.expectedAPIElbName != a.APIElbName {
				t.Errorf("expected API ELB Name, got %q, want %q", a.APIElbName, tc.expectedAPIElbName)
			}

			if !reflect.DeepEqual(tc.expectedAPIElbPortsToOpen, a.APIElbPortsToOpen) {
				t.Errorf("expected API ELB Ports To Open, got %q, want %q", a.APIElbPortsToOpen, tc.expectedAPIElbPortsToOpen)
			}

			if tc.expectedAPIElbScheme != a.APIElbScheme {
				t.Errorf("expected API ELB Scheme, got %q, want %q", a.APIElbScheme, tc.expectedAPIElbScheme)
			}

			if tc.expectedELBHealthCheckHealthyThreshold != a.ELBHealthCheckHealthyThreshold {
				t.Errorf("expected ELB health check healthy threshold, got %q, want %q", a.ELBHealthCheckHealthyThreshold, tc.expectedELBHealthCheckHealthyThreshold)
			}

			if tc.expectedELBHealthCheckInterval != a.ELBHealthCheckInterval {
				t.Errorf("expected ELB health check interval, got %q, want %q", a.ELBHealthCheckInterval, tc.expectedELBHealthCheckInterval)
			}

			if tc.expectedELBHealthCheckTimeout != a.ELBHealthCheckTimeout {
				t.Errorf("expected ELB health check timeout, got %q, want %q", a.ELBHealthCheckTimeout, tc.expectedELBHealthCheckTimeout)
			}

			if tc.expectedELBHealthCheckUnhealthyThreshold != a.ELBHealthCheckUnhealthyThreshold {
				t.Errorf("expected ELB health check unhealthy threshold, got %q, want %q", a.ELBHealthCheckUnhealthyThreshold, tc.expectedELBHealthCheckUnhealthyThreshold)
			}

			if tc.expectedIngressElbIdleTimoutSeconds != a.IngressElbIdleTimoutSeconds {
				t.Errorf("expected Ingress ELB Idle Timeout Seconds, got %q, want %q", a.IngressElbIdleTimoutSeconds, tc.expectedIngressElbIdleTimoutSeconds)
			}

			if tc.expectedIngressElbName != a.IngressElbName {
				t.Errorf("expected Ingress ELB Name, got %q, want %q", a.IngressElbName, tc.expectedIngressElbName)
			}

			if !reflect.DeepEqual(tc.expectedIngressElbPortsToOpen, a.IngressElbPortsToOpen) {
				t.Errorf("expected Ingress ELB Ports To Open, got %v, want %v", a.IngressElbPortsToOpen, tc.expectedIngressElbPortsToOpen)
			}

			if tc.expectedIngressElbScheme != a.IngressElbScheme {
				t.Errorf("expected Ingress ELB Scheme, got %q, want %q", a.IngressElbScheme, tc.expectedIngressElbScheme)
			}
		})
	}
}
