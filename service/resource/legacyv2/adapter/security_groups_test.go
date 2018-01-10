package adapter

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterSecurityGroupsRegularFields(t *testing.T) {
	testCases := []struct {
		description                       string
		customObject                      v1alpha1.AWSConfig
		expectedError                     bool
		expectedMasterSecurityGroupName   string
		expectedMasterSecurityGroupRules  []securityGroupRule
		expectedWorkerSecurityGroupName   string
		expectedWorkerSecurityGroupRules  []securityGroupRule
		expectedIngressSecurityGroupName  string
		expectedIngressSecurityGroupRules []securityGroupRule
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							PeerID: "vpc-1234",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							IngressController: v1alpha1.ClusterKubernetesIngressController{
								SecurePort:   30010,
								InsecurePort: 30011,
							},
						},
					},
				},
			},
			expectedError:                   false,
			expectedMasterSecurityGroupName: "test-cluster-master",
			expectedMasterSecurityGroupRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "0.0.0.0/0",
				},
				{
					Port:       10250,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       10300,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       22,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
			},
			expectedWorkerSecurityGroupName: "test-cluster-worker",
			expectedWorkerSecurityGroupRules: []securityGroupRule{
				{
					Port:                30010,
					Protocol:            "tcp",
					SourceSecurityGroup: "IngressSecurityGroup",
				},
				{
					Port:                30011,
					Protocol:            "tcp",
					SourceSecurityGroup: "IngressSecurityGroup",
				},
				{
					Port:       30010,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       10250,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       10300,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       22,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
			},
			expectedIngressSecurityGroupName: "test-cluster-ingress",
			expectedIngressSecurityGroupRules: []securityGroupRule{
				{
					Port:       80,
					Protocol:   "tcp",
					SourceCIDR: "0.0.0.0/0",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "0.0.0.0/0",
				},
			},
		},
	}
	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{
				vpcCIDR: "10.0.0.0/16",
			},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getSecurityGroups(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.MasterSecurityGroupName != tc.expectedMasterSecurityGroupName {
				t.Errorf("unexpected MasterGroupName, got %q, want %q", a.MasterSecurityGroupName, tc.expectedMasterSecurityGroupName)
			}

			if a.WorkerSecurityGroupName != tc.expectedWorkerSecurityGroupName {
				t.Errorf("unexpected WorkerGroupName, got %q, want %q", a.WorkerSecurityGroupName, tc.expectedWorkerSecurityGroupName)
			}

			if !reflect.DeepEqual(a.WorkerSecurityGroupRules, tc.expectedWorkerSecurityGroupRules) {
				t.Errorf("unexpected Worker Security Group Rules, got %v, want %v", a.WorkerSecurityGroupRules, tc.expectedWorkerSecurityGroupRules)
			}

			if a.IngressSecurityGroupName != tc.expectedIngressSecurityGroupName {
				t.Errorf("unexpected IngressGroupName, got %q, want %q", a.IngressSecurityGroupName, tc.expectedIngressSecurityGroupName)
			}

			if !reflect.DeepEqual(a.IngressSecurityGroupRules, tc.expectedIngressSecurityGroupRules) {
				t.Errorf("unexpected Ingress Security Group Rules, got %v, want %v", a.IngressSecurityGroupRules, tc.expectedIngressSecurityGroupRules)
			}
		})
	}
}
