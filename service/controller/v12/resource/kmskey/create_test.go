package kmskey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/client/aws"
	servicecontext "github.com/giantswarm/aws-operator/service/controller/v12/context"
)

func Test_Resource_KMSKey_newCreate(t *testing.T) {
	t.Parallel()
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
		expectedChange KMSKeyState
		description    string
	}{
		{
			description:    "current state empty, desired state empty, empty create change",
			currentState:   KMSKeyState{},
			desiredState:   KMSKeyState{},
			expectedChange: KMSKeyState{},
		},
		{
			description:  "current state empty, desired state not empty, create change == desired state",
			currentState: KMSKeyState{},
			desiredState: KMSKeyState{
				KeyAlias: "mykeyid",
			},
			expectedChange: KMSKeyState{
				KeyAlias: "mykeyid",
			},
		},
		{
			description: "current state not empty, desired state not empty, create change == desired state",
			currentState: KMSKeyState{
				KeyAlias: "currentkeyid",
			},
			desiredState: KMSKeyState{
				KeyAlias: "mykeyid",
			},
			expectedChange: KMSKeyState{
				KeyAlias: "mykeyid",
			},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	awsClients := aws.Clients{
		KMS: &KMSClientMock{},
	}
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.InstallationName = "test-install"

	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			ctx = servicecontext.NewContext(ctx, servicecontext.Context{AWSClient: awsClients})

			result, err := newResource.newCreateChange(ctx, customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			createChange, ok := result.(KMSKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}
			if createChange.KeyAlias != tc.expectedChange.KeyAlias {
				t.Errorf("expected %s, got %s", tc.expectedChange.KeyAlias, createChange.KeyAlias)
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
		createChange KMSKeyState
		description  string
	}{
		{
			description: "basic case, create",
			createChange: KMSKeyState{
				KeyAlias: "alias/test-cluster",
			},
		},
		{
			description:  "empty create change, not create",
			createChange: KMSKeyState{},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	awsClients := aws.Clients{
		KMS: &KMSClientMock{
			clusterID: "test-cluster",
		},
	}
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.InstallationName = "test-install"

	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		ctx := context.TODO()
		ctx = servicecontext.NewContext(ctx, servicecontext.Context{AWSClient: awsClients})

		err := newResource.ApplyCreateChange(ctx, customObject, tc.createChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
