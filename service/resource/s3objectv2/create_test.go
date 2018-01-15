package s3objectv2

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr/certificatetprtest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeytpr/randomkeytprtest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

func Test_Resource_S3Object_newCreate(t *testing.T) {
	clusterTpo := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description   string
		obj           v1alpha1.AWSConfig
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
				"master": BucketObjectState{
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": BucketObjectState{
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
				"mykey": BucketObjectState{
					Body:   "mykey",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": BucketObjectState{
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": BucketObjectState{
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
				"master": BucketObjectState{
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": BucketObjectState{
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": BucketObjectState{
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": BucketObjectState{},
				"worker": BucketObjectState{
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
				"master": BucketObjectState{
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": BucketObjectState{
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			desiredState: map[string]BucketObjectState{
				"master": BucketObjectState{
					Body:   "master-body",
					Bucket: "mybucket",
					Key:    "master",
				},
				"worker": BucketObjectState{
					Body:   "worker-body",
					Bucket: "mybucket",
					Key:    "worker",
				},
			},
			expectedState: map[string]BucketObjectState{
				"master": BucketObjectState{},
				"worker": BucketObjectState{},
			},
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = Clients{
		S3: &S3ClientMock{},
	}
	resourceConfig.AwsService = awsservice.AwsServiceMock{}
	resourceConfig.CertWatcher = certificatetprtest.NewService()
	resourceConfig.CloudConfig = &CloudConfigMock{}
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.RandomKeyWatcher = randomkeytprtest.NewService()
	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newCreateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
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
