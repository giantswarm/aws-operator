package adapter

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterSecurityGroupsRegularFields(t *testing.T) {
	t.Parallel()
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
					Port:       4194,
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
					Port:       10301,
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
					Port:       4194,
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
					Port:       10301,
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
		hostClients := Clients{
			EC2: &EC2ClientMock{
				vpcCIDR: "10.0.0.0/16",
			},
			IAM: &IAMClientMock{},
			STS: &STSClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      Clients{},
				HostClients:  hostClients,
			}
			err := a.Guest.SecurityGroups.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.SecurityGroups.MasterSecurityGroupName != tc.expectedMasterSecurityGroupName {
				t.Errorf("unexpected MasterGroupName, got %q, want %q", a.Guest.SecurityGroups.MasterSecurityGroupName, tc.expectedMasterSecurityGroupName)
			}

			if a.Guest.SecurityGroups.WorkerSecurityGroupName != tc.expectedWorkerSecurityGroupName {
				t.Errorf("unexpected WorkerGroupName, got %q, want %q", a.Guest.SecurityGroups.WorkerSecurityGroupName, tc.expectedWorkerSecurityGroupName)
			}

			if !reflect.DeepEqual(a.Guest.SecurityGroups.WorkerSecurityGroupRules, tc.expectedWorkerSecurityGroupRules) {
				t.Errorf("unexpected Worker Security Group Rules, got %v, want %v", a.Guest.SecurityGroups.WorkerSecurityGroupRules, tc.expectedWorkerSecurityGroupRules)
			}

			if a.Guest.SecurityGroups.IngressSecurityGroupName != tc.expectedIngressSecurityGroupName {
				t.Errorf("unexpected IngressGroupName, got %q, want %q", a.Guest.SecurityGroups.IngressSecurityGroupName, tc.expectedIngressSecurityGroupName)
			}

			if !reflect.DeepEqual(a.Guest.SecurityGroups.IngressSecurityGroupRules, tc.expectedIngressSecurityGroupRules) {
				t.Errorf("unexpected Ingress Security Group Rules, got %v, want %v", a.Guest.SecurityGroups.IngressSecurityGroupRules, tc.expectedIngressSecurityGroupRules)
			}
		})
	}
}

func TestAdapterSecurityGroupsKubernetesAPIRules(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description            string
		customObject           v1alpha1.AWSConfig
		apiWhitelistingEnabled bool
		apiWhitelistSubnets    string
		elasticIPs             []string
		hostClusterCIDR        string
		expectedError          bool
		expectedRules          []securityGroupRule
	}{
		{
			description: "case 0: API whitelisting disabled",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								SecurePort: 443,
							},
						},
					},
				},
			},
			apiWhitelistingEnabled: false,
			expectedError:          false,
			expectedRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "0.0.0.0/0",
				},
			},
		},
		{
			description: "case 1: API whitelisting enabled with default rules",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							CIDR: "10.1.1.0/24",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								SecurePort: 443,
							},
						},
					},
				},
			},
			apiWhitelistingEnabled: true,
			hostClusterCIDR:        "10.0.0.0/16",
			expectedError:          false,
			expectedRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.1.1.0/24",
				},
			},
		},
		{
			description: "case 2: API whitelisting enabled with single configured subnet",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							CIDR: "10.1.1.0/24",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								SecurePort: 443,
							},
						},
					},
				},
			},
			apiWhitelistingEnabled: true,
			apiWhitelistSubnets:    "212.145.136.84/32",
			hostClusterCIDR:        "10.0.0.0/16",
			expectedError:          false,
			expectedRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.1.1.0/24",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "212.145.136.84/32",
				},
			},
		},
		{
			description: "case 3: API whitelisting enabled with multiple configured subnets",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							CIDR: "10.1.1.0/24",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								SecurePort: 443,
							},
						},
					},
				},
			},
			apiWhitelistingEnabled: true,
			apiWhitelistSubnets:    "212.145.136.84/32,192.168.1.0/24,10.2.2.0/24",
			hostClusterCIDR:        "10.0.0.0/16",
			expectedError:          false,
			expectedRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.1.1.0/24",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "212.145.136.84/32",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "192.168.1.0/24",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.2.2.0/24",
				},
			},
		},
		{
			description: "case 4: API whitelisting enabled with NAT gateway EIPs",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							CIDR: "10.1.1.0/24",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								SecurePort: 443,
							},
						},
					},
				},
			},
			apiWhitelistingEnabled: true,
			elasticIPs: []string{
				"21.1.136.42",
				"21.2.136.84",
			},
			hostClusterCIDR: "10.0.0.0/16",
			expectedError:   false,
			expectedRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.1.1.0/24",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "21.1.136.42/32",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "21.2.136.84/32",
				},
			},
		},
		{
			description: "case 5: API whitelisting enabled with subnets and NAT gateway EIPs",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							CIDR: "10.1.1.0/24",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
						Kubernetes: v1alpha1.ClusterKubernetes{
							API: v1alpha1.ClusterKubernetesAPI{
								SecurePort: 443,
							},
						},
					},
				},
			},
			apiWhitelistingEnabled: true,
			apiWhitelistSubnets:    "212.145.136.84/32,192.168.1.1/24",
			elasticIPs: []string{
				"21.1.136.42",
				"21.2.136.84",
			},
			hostClusterCIDR: "10.0.0.0/16",
			expectedError:   false,
			expectedRules: []securityGroupRule{
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.0.0.0/16",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "10.1.1.0/24",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "212.145.136.84/32",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "192.168.1.1/24",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "21.1.136.42/32",
				},
				{
					Port:       443,
					Protocol:   "tcp",
					SourceCIDR: "21.2.136.84/32",
				},
			},
		},
	}

	for _, tc := range testCases {
		hostClients := Clients{
			EC2: &EC2ClientMock{
				elasticIPs: tc.elasticIPs,
			},
			STS: &STSClientMock{},
		}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				APIWhitelist: APIWhitelist{
					Enabled:    tc.apiWhitelistingEnabled,
					SubnetList: tc.apiWhitelistSubnets,
				},
				CustomObject: tc.customObject,
				Clients:      Clients{},
				HostClients:  hostClients,
			}

			rules, err := getKubernetesAPIRules(cfg, tc.hostClusterCIDR)
			if tc.expectedError && err == nil {
				t.Fatalf("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if len(tc.expectedRules) != len(rules) {
				t.Fatalf("expected %d master rules got %d", len(tc.expectedRules), len(rules))
			}

			if !reflect.DeepEqual(tc.expectedRules, rules) {
				t.Fatalf("expected master rules %v got %v", tc.expectedRules, rules)
			}
		})
	}
}
