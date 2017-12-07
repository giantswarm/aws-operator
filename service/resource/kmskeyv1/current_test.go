package kmskeyv1

import (
	"context"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_CurrentState(t *testing.T) {
	clusterTpo := &awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
			},
		},
	}

	testCases := []struct {
		obj              *awstpr.CustomObject
		description      string
		expectedKeyID    string
		expectedARN      string
		expectedKMSError bool
	}{
		{
			description:   "basic match",
			obj:           clusterTpo,
			expectedKeyID: "mykeyid",
			expectedARN:   "myarn",
		},
		{
			description:      "KMS error",
			obj:              clusterTpo,
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
					aRN:     tc.expectedARN,
					isError: tc.expectedKMSError,
				},
			}
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
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
				t.Errorf("expeccted keyID %q, got %q", tc.expectedKeyID, currentState.KeyID)
			}
			if currentState.ARN != tc.expectedARN {
				t.Errorf("expeccted keyID %q, got %q", tc.expectedARN, currentState.ARN)
			}
		})
	}
}
