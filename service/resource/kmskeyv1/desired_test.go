package kmskeyv1

import (
	"context"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_DesiredState(t *testing.T) {
	customObject := &awstpr.CustomObject{
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
		expectedKeyAlias string
	}{
		{
			description:      "basic match",
			obj:              customObject,
			expectedKeyAlias: "alias/test-cluster",
		},
	}
	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(KMSKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if currentState.KeyAlias != tc.expectedKeyAlias {
				t.Errorf("expected keyID %q, got %q", tc.expectedKeyAlias, currentState.KeyAlias)
			}
		})
	}
}
