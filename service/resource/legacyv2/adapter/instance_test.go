package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterInstanceRegularFields(t *testing.T) {
	testCases := []struct {
		description                    string
		customObject                   v1alpha1.AWSConfig
		errorMatcher                   func(error) bool
		expectedAZ                     string
		expectedIAMInstanceProfileName string
		expectedImageID                string
		expectedInstanceType           string
		expectedSecurityGroupID        string
		expectedSubnetID               string
	}{
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
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "eu-central-1a",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "ami-test",
								InstanceType: "m3.large",
							},
						},
					},
				},
			},
			errorMatcher:                   nil,
			expectedAZ:                     "eu-central-1a",
			expectedIAMInstanceProfileName: "test-cluster-master-EC2-K8S-Role",
			expectedImageID:                "ami-test",
			expectedInstanceType:           "m3.large",
		},
	}

	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getInstance(tc.customObject, clients)
			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if a.MasterAZ != tc.expectedAZ {
				t.Errorf("unexpected MasterAZ, got %q, want %q", a.instanceAdapter.MasterAZ, tc.expectedAZ)
			}

			if a.MasterIAMInstanceProfileName != tc.expectedIAMInstanceProfileName {
				t.Errorf("unexpected MasterIAMInstanceProfileName, got %q, want %q", a.instanceAdapter.MasterAZ, tc.expectedAZ)
			}

			if a.MasterImageID != tc.expectedImageID {
				t.Errorf("unexpected MasterImageID, got %q, want %q", a.instanceAdapter.MasterImageID, tc.expectedAZ)
			}

			if a.MasterInstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected MasterInstanceType, got %q, want %q", a.instanceAdapter.MasterInstanceType, tc.expectedInstanceType)
			}
		})
	}
}

func TestAdapterInstanceSecurityGroupID(t *testing.T) {
	testCases := []struct {
		description             string
		customObject            v1alpha1.AWSConfig
		expectedSecurityGroupID string
		expectedError           bool
		unexistingSG            bool
	}{
		{
			description: "existent security group",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedSecurityGroupID: "test-cluster",
		},
		{
			description: "unexistent security group",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			unexistingSG:            true,
			expectedError:           true,
			expectedSecurityGroupID: "",
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{
			EC2: &EC2ClientMock{
				unexistingSg: tc.unexistingSG,
				sgID:         tc.expectedSecurityGroupID,
			},
			IAM: &IAMClientMock{},
		}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getInstance(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.MasterSecurityGroupID != tc.expectedSecurityGroupID {
				t.Errorf("unexpected SecurityGroupID, got %q, want %q", a.MasterSecurityGroupID, tc.expectedSecurityGroupID)
			}
		})
	}
}

func TestAdapterInstanceSubnetID(t *testing.T) {
	testCases := []struct {
		description      string
		customObject     v1alpha1.AWSConfig
		expectedSubnetID string
		expectedError    bool
		unexistingSubnet bool
	}{
		{
			description: "existent subnet",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedSubnetID: "subnet-1234",
		},
		{
			description: "unexistent subnet",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			unexistingSubnet: true,
			expectedError:    true,
			expectedSubnetID: "",
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{
			EC2: &EC2ClientMock{
				unexistingSubnet: tc.unexistingSubnet,
				subnetID:         tc.expectedSubnetID,
			},
			IAM: &IAMClientMock{},
		}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getInstance(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.MasterSubnetID != tc.expectedSubnetID {
				t.Errorf("unexpected SubnetID, got %q, want %q", a.MasterSubnetID, tc.expectedSubnetID)
			}
		})
	}
}
