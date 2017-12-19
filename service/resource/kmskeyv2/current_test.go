package kmskeyv2

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_CurrentState(t *testing.T) {
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description      string
		expectedKeyID    string
		expectedKMSError bool
	}{
		{
			description:   "basic match",
			expectedKeyID: "mykeyid",
		},
		{
			description:      "KMS error",
			expectedKMSError: true,
		},
	}
	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resourceConfig.Clients = Clients{
				KMS: &KMSClientMock{
					keyID:   tc.expectedKeyID,
					isError: tc.expectedKMSError,
				},
			}
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), customObject)
			if err != nil && !tc.expectedKMSError {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(KMSKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if err == nil && tc.expectedKMSError {
				t.Error("expected KMS error didn't happen")
			}

			if currentState.KeyID != tc.expectedKeyID {
				t.Errorf("expected keyID %q, got %q", tc.expectedKeyID, currentState.KeyID)
			}
		})
	}
}
