package kmskeyv2

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_KMSKey_newCreate(t *testing.T) {
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		currentState   KMSKeyState
		desiredState   KMSKeyState
		expectedChange *kms.CreateKeyInput
		description    string
	}{
		{
			description:    "current state empty, desired state empty, empty create change",
			currentState:   KMSKeyState{},
			desiredState:   KMSKeyState{},
			expectedChange: nil,
		},
		{
			description:  "current state empty, desired state not empty, create change == desired state",
			currentState: KMSKeyState{},
			desiredState: KMSKeyState{
				KeyAlias: "mykeyid",
			},
			expectedChange: &kms.CreateKeyInput{},
		},
		{
			description: "current state not empty, desired state not empty, create change == desired state",
			currentState: KMSKeyState{
				KeyAlias: "currentkeyid",
			},
			desiredState: KMSKeyState{
				KeyAlias: "mykeyid",
			},
			expectedChange: &kms.CreateKeyInput{},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = Clients{
		KMS: &KMSClientMock{},
	}
	resourceConfig.Logger = microloggertest.New()
	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		result, err := newResource.newCreateChange(context.TODO(), customObject, tc.currentState, tc.desiredState)
		if err != nil {
			t.Errorf("expected '%v' got '%#v'", nil, err)
		}
		if result == nil && tc.expectedChange == nil {
			continue
		}

		createChange, ok := result.(*kms.CreateKeyInput)
		if !ok {
			t.Errorf("expected '%T', got '%T'", createChange, result)
		}
		if !reflect.DeepEqual(createChange, tc.expectedChange) {
			t.Errorf("expected change %s, got %s", tc.expectedChange, createChange)
		}
	}
}

func Test_ApplyCreateChange(t *testing.T) {
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		createChange *kms.CreateKeyInput
		description  string
	}{
		{
			description:  "basic case, create",
			createChange: &kms.CreateKeyInput{},
		},
		{
			description:  "empty create change, not create",
			createChange: nil,
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = Clients{
		KMS: &KMSClientMock{
			clusterID: "test-cluster",
		},
	}
	resourceConfig.Logger = microloggertest.New()
	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		err := newResource.ApplyCreateChange(context.TODO(), customObject, tc.createChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
