package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterWorkerPolicyRegularFields(t *testing.T) {
	testCases := []struct {
		description               string
		customObject              v1alpha1.AWSConfig
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
			err := a.getWorkerPolicy(tc.customObject, clients)
			if err != nil {
				t.Errorf("unexpected error %v", err)
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

func TestAdapterWorkerPolicyKMSKeyARN(t *testing.T) {
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
			err := a.getWorkerPolicy(tc.customObject, clients)
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

func TestAdapterWorkerPolicyS3Bucket(t *testing.T) {
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
			err := a.getWorkerPolicy(tc.customObject, clients)
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
