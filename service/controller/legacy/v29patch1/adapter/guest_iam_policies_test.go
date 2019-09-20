package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterIamPoliciesRegularFields(t *testing.T) {
	t.Parallel()
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

	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
			}
			err := a.Guest.IAMPolicies.Adapt(cfg)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.IAMPolicies.MasterPolicyName != tc.expectedMasterPolicyName {
				t.Errorf("unexpected MasterPolicyName, got %q, want %q", a.Guest.IAMPolicies.MasterPolicyName, tc.expectedMasterPolicyName)
			}

			if a.Guest.IAMPolicies.MasterRoleName != tc.expectedMasterRoleName {
				t.Errorf("unexpected MasterRoleName, got %q, want %q", a.Guest.IAMPolicies.MasterRoleName, tc.expectedMasterRoleName)
			}

			if a.Guest.IAMPolicies.MasterProfileName != tc.expectedMasterProfileName {
				t.Errorf("unexpected MasterProfileName, got %q, want %q", a.Guest.IAMPolicies.MasterProfileName, tc.expectedMasterProfileName)
			}

			if a.Guest.IAMPolicies.WorkerPolicyName != tc.expectedWorkerPolicyName {
				t.Errorf("unexpected WorkerPolicyName, got %q, want %q", a.Guest.IAMPolicies.WorkerPolicyName, tc.expectedWorkerPolicyName)
			}

			if a.Guest.IAMPolicies.WorkerRoleName != tc.expectedWorkerRoleName {
				t.Errorf("unexpected WorkerRoleName, got %q, want %q", a.Guest.IAMPolicies.WorkerRoleName, tc.expectedWorkerRoleName)
			}

			if a.Guest.IAMPolicies.WorkerProfileName != tc.expectedWorkerProfileName {
				t.Errorf("unexpected WorkerProfileName, got %q, want %q", a.Guest.IAMPolicies.WorkerProfileName, tc.expectedWorkerProfileName)
			}
		})
	}
}

func TestAdapterIamPoliciesS3Bucket(t *testing.T) {
	t.Parallel()
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
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject:           tc.customObject,
				TenantClusterAccountID: tc.accountID,
			}
			err := a.Guest.IAMPolicies.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Errorf("expected error didn't happen")
			}

			if !tc.expectedError {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				if a.Guest.IAMPolicies.S3Bucket != tc.expectedS3Bucket {
					t.Errorf("unexpected S3Bucket, got %q, want %q", a.Guest.IAMPolicies.S3Bucket, tc.expectedS3Bucket)
				}
			}
		})
	}
}
