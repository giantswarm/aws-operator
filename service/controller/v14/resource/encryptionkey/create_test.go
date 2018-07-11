package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
)

func Test_Resource_EncryptionKey_newCreate(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		currentState   encrypter.EncryptionKeyState
		desiredState   encrypter.EncryptionKeyState
		expectedChange encrypter.EncryptionKeyState
		description    string
	}{
		{
			description:    "current state empty, desired state empty, empty create change",
			currentState:   encrypter.EncryptionKeyState{},
			desiredState:   encrypter.EncryptionKeyState{},
			expectedChange: encrypter.EncryptionKeyState{},
		},
		{
			description:  "current state empty, desired state not empty, create change == desired state",
			currentState: encrypter.EncryptionKeyState{},
			desiredState: encrypter.EncryptionKeyState{
				KeyName: "mykeyid",
			},
			expectedChange: encrypter.EncryptionKeyState{
				KeyName: "mykeyid",
			},
		},
		{
			description: "current state not empty, desired state not empty, create change == desired state",
			currentState: encrypter.EncryptionKeyState{
				KeyName: "currentkeyid",
			},
			desiredState: encrypter.EncryptionKeyState{
				KeyName: "mykeyid",
			},
			expectedChange: encrypter.EncryptionKeyState{
				KeyName: "mykeyid",
			},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.Encrypter = &encrypter.EncrypterMock{}
	resourceConfig.InstallationName = "test-install"

	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			result, err := newResource.newCreateChange(ctx, customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			createChange, ok := result.(encrypter.EncryptionKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}
			if createChange.KeyName != tc.expectedChange.KeyName {
				t.Errorf("expected %s, got %s", tc.expectedChange.KeyName, createChange.KeyName)
			}
		})
	}
}

func Test_ApplyCreateChange(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		createChange encrypter.EncryptionKeyState
		description  string
	}{
		{
			description: "basic case, create",
			createChange: encrypter.EncryptionKeyState{
				KeyName: "alias/test-cluster",
			},
		},
		{
			description:  "empty create change, not create",
			createChange: encrypter.EncryptionKeyState{},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.Encrypter = &encrypter.EncrypterMock{}
	resourceConfig.InstallationName = "test-install"

	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		ctx := context.TODO()
		err := newResource.ApplyCreateChange(ctx, customObject, tc.createChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
