package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
)

func Test_newDelete(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		currentState  encrypter.EncryptionKeyState
		desiredState  encrypter.EncryptionKeyState
		expectedAlias string
		description   string
	}{
		{
			description:   "current and desired state empty, expected empty",
			currentState:  encrypter.EncryptionKeyState{},
			desiredState:  encrypter.EncryptionKeyState{},
			expectedAlias: "",
		},
		{
			description:  "current state empty, desired state not empty, expected empty",
			currentState: encrypter.EncryptionKeyState{},
			desiredState: encrypter.EncryptionKeyState{
				KeyName: "desired",
			},
			expectedAlias: "",
		},
		{
			description: "current state not empty, desired state not empty but equal, expected current state",
			currentState: encrypter.EncryptionKeyState{
				KeyName: "current",
			},
			desiredState: encrypter.EncryptionKeyState{
				KeyName: "current",
			},
			expectedAlias: "current",
		},
	}

	var err error
	var newResource *Resource
	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.Encrypter = &EncrypterMock{}
	resourceConfig.InstallationName = "test-install"

	newResource, err = New(resourceConfig)
	if err != nil {
		t.Error("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newDeleteChange(context.TODO(), customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			deleteChange, ok := result.(encrypter.EncryptionKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteChange, result)
			}
			if deleteChange.KeyName != tc.expectedAlias {
				t.Errorf("expected %s, got %s", tc.expectedAlias, deleteChange.KeyName)
			}
		})
	}
}

func Test_ApplyDeleteChange(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		deleteChange encrypter.EncryptionKeyState
		description  string
	}{
		{
			description: "basic case, create",
			deleteChange: encrypter.EncryptionKeyState{
				KeyName: "alias/test-cluster",
			},
		},
		{
			description:  "empty create change, not create",
			deleteChange: encrypter.EncryptionKeyState{},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.Encrypter = &EncrypterMock{}
	resourceConfig.InstallationName = "test-install"

	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		ctx := context.TODO()
		err := newResource.ApplyDeleteChange(ctx, &customObject, tc.deleteChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
