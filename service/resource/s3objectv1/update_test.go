package s3objectv1

import (
	"bufio"
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr/certificatetprtest"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_S3Object_newUpdate(t *testing.T) {
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
		obj            awstpr.CustomObject
		currentState   BucketObjectState
		desiredState   BucketObjectState
		expectedBody   string
		expectedBucket string
		expectedKey    string
		description    string
	}{
		{
			description:    "current state empty, desired state empty, empty create change",
			obj:            clusterTpo,
			currentState:   BucketObjectState{},
			desiredState:   BucketObjectState{},
			expectedBody:   "",
			expectedBucket: "",
			expectedKey:    "",
		},
		{
			description:  "current state empty, desired state not empty, create change == desired state",
			obj:          clusterTpo,
			currentState: BucketObjectState{},
			desiredState: BucketObjectState{
				WorkerCloudConfig: BucketObjectInstance{
					Body:   "mybody",
					Bucket: "mybucket",
					Key:    "mykey",
				},
			},
			expectedBody:   "mybody",
			expectedBucket: "mybucket",
			expectedKey:    "mykey",
		},
		{
			description: "current state not empty, desired state not empty, create change == desired state",
			obj:         clusterTpo,
			currentState: BucketObjectState{
				WorkerCloudConfig: BucketObjectInstance{
					Body:   "currentbody",
					Bucket: "currentbucket",
					Key:    "currentkey",
				},
			},
			desiredState: BucketObjectState{
				WorkerCloudConfig: BucketObjectInstance{
					Body:   "mybody",
					Bucket: "mybucket",
					Key:    "mykey",
				},
			},
			expectedBody:   "mybody",
			expectedBucket: "mybucket",
			expectedKey:    "mykey",
		},
	}

	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = Clients{
		S3: &S3ClientMock{},
	}
	resourceConfig.AwsService = awsservice.AwsServiceMock{}
	resourceConfig.CertWatcher = &certificatetprtest.Service{}
	resourceConfig.CloudConfig = &CloudConfigMock{}
	resourceConfig.Logger = microloggertest.New()
	newResource, err = New(resourceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			createChange, ok := result.(s3.PutObjectInput)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}
			if createChange.Key != nil && *createChange.Key != tc.expectedKey {
				t.Errorf("expected key %s, got %s", tc.expectedKey, createChange.Key)
			}
			if createChange.Bucket != nil && *createChange.Bucket != tc.expectedBucket {
				t.Errorf("expected bucket %s, got %s", tc.expectedBucket, createChange.Bucket)
			}

			if createChange.Body != nil {
				var actualBodyItems []string
				scanner := bufio.NewScanner(createChange.Body)
				for scanner.Scan() {
					actualBodyItems = append(actualBodyItems, scanner.Text())
				}
				actualBody := strings.Join(actualBodyItems, "\n")
				if actualBody != tc.expectedBody {
					t.Errorf("expected body %s, got %s", tc.expectedBody, actualBody)
				}
			}
		})
	}
}
