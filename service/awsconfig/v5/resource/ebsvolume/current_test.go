package ebsvolume

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_CurrentState(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description   string
		obj           *v1alpha1.AWSConfig
		expectedState *EBSVolumeState
		ebsVolumes    []EBSVolumeMock
	}{
		{
			description: "basic match with no ebs volumes",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				VolumeIDs: []string{},
			},
		},
		{
			description: "basic match with ebs volume",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				VolumeIDs: []string{
					"vol-1234",
				},
			},
			ebsVolumes: []EBSVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
			},
		},
		{
			description: "basic match with multiple ebs volumes",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				VolumeIDs: []string{
					"vol-1234",
					"vol-5678",
				},
			},
			ebsVolumes: []EBSVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					volumeID: "vol-5678",
					tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-5678"),
						},
					},
				},
			},
		},
		{
			description: "no match due to cluster tag",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				VolumeIDs: []string{},
			},
			ebsVolumes: []EBSVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/cluster/other-cluster"),
							Value: aws.String("owned"),
						},
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
			},
		},
		{
			description: "no match due to missing pvc tag",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				VolumeIDs: []string{},
			},
			ebsVolumes: []EBSVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
					},
				},
			},
		},
	}
	var err error
	var newResource *Resource

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := Config{
				Clients: Clients{
					EC2: &EC2ClientMock{
						customObject: *tc.obj,
						ebsVolumes:   tc.ebsVolumes,
					},
				},
				Logger: microloggertest.New(),
			}
			newResource, err = New(c)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(*EBSVolumeState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if !reflect.DeepEqual(currentState, tc.expectedState) {
				t.Errorf("expected current state '%#v', got '%#v'", tc.expectedState, currentState)
			}
		})
	}
}
