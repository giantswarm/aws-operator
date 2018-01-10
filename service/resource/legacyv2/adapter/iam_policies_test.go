package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterIamPoliciesRegularFields(t *testing.T) {
	testCases := []struct {
		description               string
		customObject              v1alpha1.AWSConfig
		expectedMasterRoleName    string
		expectedMasterPolicyName  string
		expectedMasterProfileName string
		expectedWorkerRoleName    string
		expectedWorkerPolicyName  string
		expectedWorkerProfileName string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			expectedMasterRoleName:    "test-cluster-master-EC2-K8S-Role",
			expectedMasterPolicyName:  "test-cluster-master-EC2-K8S-Policy",
			expectedMasterProfileName: "test-cluster-master-EC2-K8S-Role",
			expectedWorkerRoleName:    "test-cluster-worker-EC2-K8S-Role",
			expectedWorkerPolicyName:  "test-cluster-worker-EC2-K8S-Policy",
			expectedWorkerProfileName: "test-cluster-worker-EC2-K8S-Role",
		},
	}

	clients := Clients{
		KMS: &KMSClientMock{},
		IAM: &IAMClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getIamPolicies(cfg)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.MasterPolicyName != tc.expectedMasterPolicyName {
				t.Errorf("unexpected MasterPolicyName, got %q, want %q", a.MasterPolicyName, tc.expectedMasterPolicyName)
			}

			if a.MasterRoleName != tc.expectedMasterRoleName {
				t.Errorf("unexpected MasterRoleName, got %q, want %q", a.MasterRoleName, tc.expectedMasterRoleName)
			}

			if a.MasterProfileName != tc.expectedMasterProfileName {
				t.Errorf("unexpected MasterProfileName, got %q, want %q", a.MasterProfileName, tc.expectedMasterProfileName)
			}

			if a.WorkerPolicyName != tc.expectedWorkerPolicyName {
				t.Errorf("unexpected WorkerPolicyName, got %q, want %q", a.WorkerPolicyName, tc.expectedWorkerPolicyName)
			}

			if a.WorkerRoleName != tc.expectedWorkerRoleName {
				t.Errorf("unexpected WorkerRoleName, got %q, want %q", a.WorkerRoleName, tc.expectedWorkerRoleName)
			}

			if a.WorkerProfileName != tc.expectedWorkerProfileName {
				t.Errorf("unexpected WorkerProfileName, got %q, want %q", a.WorkerProfileName, tc.expectedWorkerProfileName)
			}
		})
	}
}

func TestAdapterIamPoliciesKMSKeyARN(t *testing.T) {
	testCases := []struct {
		description       string
		customObject      v1alpha1.AWSConfig
		expectedKMSKeyARN string
		expectedError     bool
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			expectedKMSKeyARN: "alias/test-cluster",
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			clients := Clients{
				IAM: &IAMClientMock{},
				KMS: &KMSClientMock{
					keyARN:  tc.expectedKMSKeyARN,
					isError: tc.expectedError,
				},
			}
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getIamPolicies(cfg)
			if tc.expectedError && err == nil {
				t.Errorf("expected error didn't happen")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				if a.KMSKeyARN != tc.expectedKMSKeyARN {
					t.Errorf("unexpected KMSKeyARN, got %q, want %q", a.KMSKeyARN, tc.expectedKMSKeyARN)
				}
			}
		})
	}
}

func TestAdapterIamPoliciesS3Bucket(t *testing.T) {
	testCases := []struct {
		description      string
		customObject     v1alpha1.AWSConfig
		accountID        string
		expectedS3Bucket string
		expectedError    bool
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			accountID:        "111111111111",
			expectedS3Bucket: "111111111111-g8s-test-cluster",
		},
		{
			description: "client error",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			clients := Clients{
				KMS: &KMSClientMock{},
				IAM: &IAMClientMock{
					accountID: tc.accountID,
					isError:   tc.expectedError,
				},
			}
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getIamPolicies(cfg)
			if tc.expectedError && err == nil {
				t.Errorf("expected error didn't happen")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				if a.S3Bucket != tc.expectedS3Bucket {
					t.Errorf("unexpected S3Bucket, got %q, want %q", a.S3Bucket, tc.expectedS3Bucket)
				}
			}
		})
	}
}
