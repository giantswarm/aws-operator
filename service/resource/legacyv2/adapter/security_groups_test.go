package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterSecurityGroupsRegularFields(t *testing.T) {
	testCases := []struct {
		description              string
		customObject             v1alpha1.AWSConfig
		expectedError            bool
		expectedMasterGroupName  string
		expectedWorkerGroupName  string
		expectedIngressGroupName string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
			},
			expectedError:            false,
			expectedMasterGroupName:  "test-cluster-master",
			expectedWorkerGroupName:  "test-cluster-worker",
			expectedIngressGroupName: "test-cluster-ingress",
		},
	}
	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getSecurityGroups(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.MasterGroupName != tc.expectedMasterGroupName {
				t.Errorf("unexpected MasterGroupName, got %q, want %q", a.MasterGroupName, tc.expectedMasterGroupName)
			}

			if a.WorkerGroupName != tc.expectedWorkerGroupName {
				t.Errorf("unexpected WorkerGroupName, got %q, want %q", a.WorkerGroupName, tc.expectedWorkerGroupName)
			}

			if a.IngressGroupName != tc.expectedIngressGroupName {
				t.Errorf("unexpected IngressGroupName, got %q, want %q", a.IngressGroupName, tc.expectedIngressGroupName)
			}
		})
	}
}
