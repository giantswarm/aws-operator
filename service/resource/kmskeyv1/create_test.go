package kmskeyv1

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_S3Object_newCreate(t *testing.T) {
	clusterTpo := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
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
		result, err := newResource.newCreateChange(context.TODO(), clusterTpo, tc.currentState, tc.desiredState)
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
	clusterTpo := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
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
		err := newResource.ApplyCreateChange(context.TODO(), &clusterTpo, tc.createChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
