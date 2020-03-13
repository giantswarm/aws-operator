package s3object

import (
	"context"
	"reflect"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func Test_Resource_S3Object_newCreate(t *testing.T) {
	t.Parallel()
	clusterTpo := providerv1alpha1.AWSConfig{
		Spec: providerv1alpha1.AWSConfigSpec{
			Cluster: providerv1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description   string
		obj           providerv1alpha1.AWSConfig
		currentState  map[string]BucketObjectState
		desiredState  map[string]BucketObjectState
		expectedState map[string]BucketObjectState
	}{
		{
			description:   "current state empty, desired state empty, empty create change",
			obj:           clusterTpo,
			currentState:  map[string]BucketObjectState{},
			desiredState:  map[string]BucketObjectState{},
			expectedState: map[string]BucketObjectState{},
		},
		{
			description:  "current state empty, desired state not empty, create change == desired state",
			obj:          clusterTpo,
			currentState: map[string]BucketObjectState{},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
		},
		{
			description: "current state not empty, desired state not empty, create change == desired state",
			obj:         clusterTpo,
			currentState: map[string]BucketObjectState{
				"mykey": {
					Body:   "mykey",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
		},
		{
			description: "current state has 1 object, desired state has 2 objects, create change == missing object",
			obj:         clusterTpo,
			currentState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
		},
		{
			description: "current state matches desired state, empty create change",
			obj:         clusterTpo,
			currentState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": {
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {},
				"worker": {},
			},
		},
	}

	awsClients := aws.Clients{
		S3: &S3ClientMock{},
	}
	cloudConfig := &CloudConfigMock{}

	release := &releasev1alpha1.Release{}
	clientset := fake.NewSimpleClientset(release)

	var err error
	var newResource *Resource
	{
		c := Config{
			CertsSearcher:      certstest.NewSearcher(certstest.Config{}),
			CloudConfig:        cloudConfig,
			G8sClient:          clientset,
			Logger:             microloggertest.New(),
			RandomKeysSearcher: randomkeystest.NewSearcher(),
			RegistryDomain:     "example.com",
		}

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			cc := controllercontext.Context{
				Client: controllercontext.ContextClient{
					TenantCluster: controllercontext.ContextClientTenantCluster{
						AWS: awsClients,
					},
				},
			}
			ctx = controllercontext.NewContext(ctx, cc)

			result, err := newResource.newCreateChange(ctx, tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			createChange, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}

			if !reflect.DeepEqual(tc.expectedState, createChange) {
				t.Error("expected", tc.expectedState, "got", createChange)
			}
		})
	}
}
