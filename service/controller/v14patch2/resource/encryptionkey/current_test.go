package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v14patch2/encrypter"
)

func Test_CurrentState(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description             string
		expectedKeyID           string
		expectedEncryptionError bool
	}{
		{
			description:   "basic match",
			expectedKeyID: "mykeyid",
		},
		{
			description:             "Encryption error",
			expectedEncryptionError: true,
		},
	}
	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.InstallationName = "test-install"

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resourceConfig.Encrypter = &encrypter.EncrypterMock{
				IsError: tc.expectedEncryptionError,
				KeyID:   tc.expectedKeyID,
			}
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			ctx := context.TODO()
			result, err := newResource.GetCurrentState(ctx, customObject)
			if err != nil && !tc.expectedEncryptionError {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(encrypter.EncryptionKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if err == nil && tc.expectedEncryptionError {
				t.Error("expected encryption error didn't happen")
			}

			if currentState.KeyID != tc.expectedKeyID {
				t.Errorf("expected keyID %q, got %q", tc.expectedKeyID, currentState.KeyID)
			}
		})
	}
}
