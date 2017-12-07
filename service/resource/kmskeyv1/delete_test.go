package kmskeyv1

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_newDelete(t *testing.T) {
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
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
			description: "current state not empty, desired state not empty but different, expected current state",
			currentState: KMSKeyState{
				KeyAlias: "current",
			},
			desiredState: KMSKeyState{
				KeyAlias: "desired",
			},
			expectedAlias: "current",
		},
		{
			description: "current state not empty, desired state not empty but equal, expected desired state",
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
	resourceConfig.Clients = Clients{}
	resourceConfig.Logger = microloggertest.New()
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
			deleteChange, ok := result.(kms.DeleteAliasInput)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteChange, result)
			}
			if deleteChange.AliasName != nil && *deleteChange.AliasName != tc.expectedAlias {
				t.Errorf("expected %s, got %s", tc.expectedAlias, *deleteChange.AliasName)
			}
		})
	}
}

func Test_ApplyDeleteChange(t *testing.T) {
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
			},
		},
	}

	testCases := []struct {
		deleteChange kms.DeleteAliasInput
		description  string
	}{
		{
			description: "basic case, create",
			deleteChange: kms.DeleteAliasInput{
				AliasName: aws.String("alias/test-cluster"),
			},
		},
		{
			description:  "empty create change, not create",
			deleteChange: kms.DeleteAliasInput{},
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
		err := newResource.ApplyDeleteChange(context.TODO(), &customObject, tc.deleteChange)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
}
