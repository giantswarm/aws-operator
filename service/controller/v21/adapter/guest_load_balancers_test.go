package adapter

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v21/key"
)

func TestAdapterLoadBalancersRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description                              string
		customObject                             v1alpha1.AWSConfig
		errorMatcher                             func(error) bool
		expectedAPIElbName                       string
		expectedAPIElbPortsToOpen                []GuestLoadBalancersAdapterPortPair
		expectedAPIElbScheme                     string
		expectedAPIElbSecurityGroupID            string
		expectedAPIElbSubnetID                   string
		expectedEtcdElbName                      string
		expectedEtcdElbPortsToOpen               []GuestLoadBalancersAdapterPortPair
		expectedEtcdElbScheme                    string
		expectedEtcdElbSecurityGroupID           string
		expectedEtcdElbSubnetID                  string
		expectedELBAZ                            string
		expectedELBHealthCheckHealthyThreshold   int
		expectedELBHealthCheckInterval           int
		expectedELBHealthCheckTimeout            int
		expectedELBHealthCheckUnhealthyThreshold int
		expectedIngressElbName                   string
		expectedIngressElbPortsToOpen            []GuestLoadBalancersAdapterPortPair
		expectedIngressElbScheme                 string
	}{
		{
			description: "empty custom object with AZs (to test for missing cloud config key",
			customObject: v1alpha1.AWSConfig{
				Status: v1alpha1.AWSConfigStatus{
					AWS: v1alpha1.AWSConfigStatusAWS{
						AvailabilityZones: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
							{
								Name: "eu-central-1a",
							},
						},
					},
				},
			},
			errorMatcher: key.IsMissingCloudConfigKey,
		},
		{
			description:  "empty custom object",
			customObject: v1alpha1.AWSConfig{},
			errorMatcher: IsInvalidConfig,
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Etcd: v1alpha1.ClusterEtcd{
							Domain: "etcd.test-cluster.aws.giantswarm.io",
							Port:   2379,
						},
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
						AZ: "eu-central-1a",
					},
				},
				Status: v1alpha1.AWSConfigStatus{
					AWS: v1alpha1.AWSConfigStatusAWS{
						AvailabilityZones: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
							{
								Name: "eu-central-1a",
							},
						},
					},
				},
			},
			errorMatcher:       nil,
			expectedAPIElbName: "test-cluster-api",
			expectedAPIElbPortsToOpen: []GuestLoadBalancersAdapterPortPair{
				{
					PortELB:      443,
					PortInstance: 443,
				},
			},
			expectedAPIElbScheme: "internet-facing",
			expectedEtcdElbName:  "test-cluster-etcd",
			expectedEtcdElbPortsToOpen: []GuestLoadBalancersAdapterPortPair{
				{
					PortELB:      2379,
					PortInstance: 2379,
				},
			},
			expectedEtcdElbScheme:                    "internal",
			expectedELBAZ:                            "eu-central-1a",
			expectedELBHealthCheckHealthyThreshold:   2,
			expectedELBHealthCheckInterval:           5,
			expectedELBHealthCheckTimeout:            3,
			expectedELBHealthCheckUnhealthyThreshold: 2,
			expectedIngressElbName:                   "test-cluster-ingress",
			expectedIngressElbPortsToOpen: []GuestLoadBalancersAdapterPortPair{
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
			STS: &STSClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.Guest.LoadBalancers.Adapt(cfg)

			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if tc.expectedAPIElbName != a.Guest.LoadBalancers.APIElbName {
				t.Errorf("expected API ELB Name, got %q, want %q", a.Guest.LoadBalancers.APIElbName, tc.expectedAPIElbName)
			}

			if !reflect.DeepEqual(tc.expectedAPIElbPortsToOpen, a.Guest.LoadBalancers.APIElbPortsToOpen) {
				t.Errorf("expected API ELB Ports To Open, got %q, want %q", a.Guest.LoadBalancers.APIElbPortsToOpen, tc.expectedAPIElbPortsToOpen)
			}

			if tc.expectedAPIElbScheme != a.Guest.LoadBalancers.APIElbScheme {
				t.Errorf("expected API ELB Scheme, got %q, want %q", a.Guest.LoadBalancers.APIElbScheme, tc.expectedAPIElbScheme)
			}

			if tc.expectedEtcdElbName != a.Guest.LoadBalancers.EtcdElbName {
				t.Errorf("expected etcd ELB Name, got %q, want %q", a.Guest.LoadBalancers.EtcdElbName, tc.expectedEtcdElbName)
			}

			if !reflect.DeepEqual(tc.expectedEtcdElbPortsToOpen, a.Guest.LoadBalancers.EtcdElbPortsToOpen) {
				t.Errorf("expected etcd ELB Ports To Open, got %q, want %q", a.Guest.LoadBalancers.EtcdElbPortsToOpen, tc.expectedEtcdElbPortsToOpen)
			}

			if tc.expectedEtcdElbScheme != a.Guest.LoadBalancers.EtcdElbScheme {
				t.Errorf("expected etcd ELB Scheme, got %q, want %q", a.Guest.LoadBalancers.EtcdElbScheme, tc.expectedEtcdElbScheme)
			}

			if tc.expectedELBHealthCheckHealthyThreshold != a.Guest.LoadBalancers.ELBHealthCheckHealthyThreshold {
				t.Errorf("expected ELB health check healthy threshold, got %q, want %q", a.Guest.LoadBalancers.ELBHealthCheckHealthyThreshold, tc.expectedELBHealthCheckHealthyThreshold)
			}

			if tc.expectedELBHealthCheckInterval != a.Guest.LoadBalancers.ELBHealthCheckInterval {
				t.Errorf("expected ELB health check interval, got %q, want %q", a.Guest.LoadBalancers.ELBHealthCheckInterval, tc.expectedELBHealthCheckInterval)
			}

			if tc.expectedELBHealthCheckTimeout != a.Guest.LoadBalancers.ELBHealthCheckTimeout {
				t.Errorf("expected ELB health check timeout, got %q, want %q", a.Guest.LoadBalancers.ELBHealthCheckTimeout, tc.expectedELBHealthCheckTimeout)
			}

			if tc.expectedELBHealthCheckUnhealthyThreshold != a.Guest.LoadBalancers.ELBHealthCheckUnhealthyThreshold {
				t.Errorf("expected ELB health check unhealthy threshold, got %q, want %q", a.Guest.LoadBalancers.ELBHealthCheckUnhealthyThreshold, tc.expectedELBHealthCheckUnhealthyThreshold)
			}

			if tc.expectedIngressElbName != a.Guest.LoadBalancers.IngressElbName {
				t.Errorf("expected Ingress ELB Name, got %q, want %q", a.Guest.LoadBalancers.IngressElbName, tc.expectedIngressElbName)
			}

			if !reflect.DeepEqual(tc.expectedIngressElbPortsToOpen, a.Guest.LoadBalancers.IngressElbPortsToOpen) {
				t.Errorf("expected Ingress ELB Ports To Open, got %v, want %v", a.Guest.LoadBalancers.IngressElbPortsToOpen, tc.expectedIngressElbPortsToOpen)
			}

			if tc.expectedIngressElbScheme != a.Guest.LoadBalancers.IngressElbScheme {
				t.Errorf("expected Ingress ELB Scheme, got %q, want %q", a.Guest.LoadBalancers.IngressElbScheme, tc.expectedIngressElbScheme)
			}
		})
	}
}
