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

func Test_Resource_S3Object_newUpdate(t *testing.T) {
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
			description:   "current state empty, desired state empty, empty update change",
			obj:           clusterTpo,
			currentState:  map[string]BucketObjectState{},
			desiredState:  map[string]BucketObjectState{},
			expectedState: map[string]BucketObjectState{},
		},
		{
			description:  "current state empty, desired state not empty, empty update change",
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
				"master": {},
			},
		},
		{
			description: "current state matches desired state, empty update change",
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
			},
			expectedState: map[string]BucketObjectState{
				"master": {},
			},
		},
		{
			description: "current state does not match desired state, update bucket object",
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
					Body:   "master-new-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": {
					Body:   "master-new-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
		},
		{
			description: "current state does not match desired state, update bucket object",
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
					Body:   "master-new-body",
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
				"master": {
					Body:   "master-new-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": {},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
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

			ctx := context.TODO()
			cc := controllercontext.Context{
				Client: controllercontext.ContextClient{
					TenantCluster: controllercontext.ContextClientTenantCluster{
						AWS: awsClients,
					},
				},
			}
			ctx = controllercontext.NewContext(ctx, cc)

			result, err := newResource.newUpdateChange(ctx, tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			updateChange, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", updateChange, result)
			}

			if !reflect.DeepEqual(tc.expectedState, updateChange) {
				t.Error("expected", tc.expectedState, "got", updateChange)
			}
		})
	}
}
