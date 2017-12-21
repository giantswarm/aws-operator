package s3objectv2

import (
	"bufio"
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
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
		obj            v1alpha1.AWSConfig
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
