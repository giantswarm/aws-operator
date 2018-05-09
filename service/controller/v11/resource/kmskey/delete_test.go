package kmskey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/client/aws"
	awsclientcontext "github.com/giantswarm/aws-operator/service/controller/v11/context/awsclient"
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
		currentState  KMSKeyState
		desiredState  KMSKeyState
		expectedAlias string
		description   string
	}{
		{
			description:   "current and desired state empty, expected empty",
			currentState:  KMSKeyState{},
			desiredState:  KMSKeyState{},
			expectedAlias: "",
		},
		{
			description:  "current state empty, desired state not empty, expected empty",
			currentState: KMSKeyState{},
			desiredState: KMSKeyState{
				KeyAlias: "desired",
			},
			expectedAlias: "",
		},
		{
			description: "current state not empty, desired state not empty but equal, expected current state",
			currentState: KMSKeyState{
				KeyAlias: "current",
			},
			desiredState: KMSKeyState{
				KeyAlias: "current",
			},
			expectedAlias: "current",
		},
	}

	var err error
	var newResource *Resource
	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
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
			deleteChange, ok := result.(KMSKeyState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteChange, result)
			}
			if deleteChange.KeyAlias != tc.expectedAlias {
				t.Errorf("expected %s, got %s", tc.expectedAlias, deleteChange.KeyAlias)
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
		deleteChange KMSKeyState
		description  string
	}{
		{
			description: "basic case, create",
			deleteChange: KMSKeyState{
				KeyAlias: "alias/test-cluster",
			},
		},
		{
			description:  "empty create change, not create",
			deleteChange: KMSKeyState{},
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
		err := newResource.ApplyDeleteChange(awsclientcontext.NewContext(context.TODO(), awsClients), &customObject, tc.deleteChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
