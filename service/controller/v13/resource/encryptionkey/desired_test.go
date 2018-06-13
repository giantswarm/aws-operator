package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
)

func Test_DesiredState(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description     string
		expectedKeyName string
	}{
		{
			description:     "basic match",
			expectedKeyName: "alias/test-cluster",
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
				KeyName: tc.expectedKeyName,
			}

			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetDesiredState(context.TODO(), customObject)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(encrypter.EncryptionKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if currentState.KeyName != tc.expectedKeyName {
				t.Errorf("expected keyName %q, got %q", tc.expectedKeyName, currentState.KeyName)
			}
		})
	}
}
